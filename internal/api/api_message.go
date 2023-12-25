package api

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"net/http"
	"net/url"
	"strconv"

	"go-form-hub/internal/model"
	"go-form-hub/internal/services/message"

	"github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type websocketClient struct {
	Conn 	*websocket.Conn
	mu   	sync.Mutex
    UserID  int64
}

var WebsocketClients = make(map[int64]*websocketClient)

func (c *websocketClient) NotifyNewMessage(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Error().Msgf("websocket write error:  %s", err)
		return
	}
}

type MessageAPIController struct {
	service         message.Service
	validator       *validator.Validate
	responseEncoder ResponseEncoder
}

func NewMessageAPIController(service message.Service, v *validator.Validate, responseEncoder ResponseEncoder) Router {
	return &MessageAPIController{
		service:         service,
		validator:       v,
		responseEncoder: responseEncoder,
	}
}

func (c *MessageAPIController) Routes() []Route {
	return []Route{
		{
			Name:         "MessageSend",
			Method:       http.MethodPost,
			Path:         "/message/send",
			Handler:      c.MessageSend,
			AuthRequired: true,
		},
		{
			Name:         "MessageCheckUnread",
			Method:       http.MethodGet,
			Path:         "/message/check",
			Handler:      c.MessageCheckUnread,
			AuthRequired: true,
		},
		{
			Name:         "GetChatsList",
			Method:       http.MethodGet,
			Path:         "/message/chats",
			Handler:      c.GetChatList,
			AuthRequired: true,
		},
		{
			Name:         "GetChatByID",
			Method:       http.MethodGet,
			Path:         "/message/chats/{id}",
			Handler:      c.FindChatByID,
			AuthRequired: true,
		},
		{
			Name:         "StartWebsocket",
			Method:       http.MethodGet,
			Path:         "/message/subscribe",
			Handler:      c.MessageStartWebsocket,
			AuthRequired: true,
		},
	}
}

func (c *MessageAPIController) MessageSend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()
	if err != nil {
		log.Error().Msgf("message_api mess_save body read error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var message model.MessageSave
	if err = json.Unmarshal(requestJSON, &message); err != nil {
		log.Error().Msgf("message_api mess_save unmarshal error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.MessageSave(r.Context(), &message)
	if err != nil {
		log.Error().Msgf("message_api mess_save error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	err = WebsocketClients[message.ReceiverID].Conn.WriteMessage(websocket.TextMessage, []byte("New Message Received!"))
	if err != nil {
		log.Error().Msgf("Websocket send message error: %s", err)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

// Этот метод проверяет, есть ли в бд записи с reciever_id = currentUser и is_read = false
// Возвращает Json формата model.CheckUnreadMessages
func (c *MessageAPIController) MessageCheckUnread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := c.service.MessageCheck(ctx)
	if err != nil {
		log.Error().Msgf("message_api mess_check error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

// Этот метод возвращает список чатов с reciever_id = currentUser или sender_id = currentUser
// Возвращает Json формата model.ChatList
func (c *MessageAPIController) GetChatList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	result, err := c.service.ChatList(ctx)
	if err != nil {
		log.Error().Msgf("message_api chat_list error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

// Этот метод возвращает чат с currentUser и пользователем с id = url
// Возвращает Json формата model.Chat
func (c *MessageAPIController) FindChatByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Msgf("message_api chat_get unescape error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = fmt.Errorf("message_api chat_get parse_id error: %v", err)
		log.Error().Msg(err.Error())
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.ChatGet(ctx, id)
	if err != nil {
		log.Error().Msgf("message_api chat_get error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

// Этот метод начинает соединение Websocket для обновления новых сообщений
func (c *MessageAPIController) MessageStartWebsocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

	upgrader.CheckOrigin = func(r *http.Request) bool {
        return true
    }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Msgf("Websocket start connection error: %s", err)
		return
	}

	client := &websocketClient{Conn: conn, UserID: currentUser.ID}
	WebsocketClients[currentUser.ID] = client
	// defer func() {
	// 	conn.Close()
	// 	delete(WebsocketClients, currentUser.ID)
	// }()

	err = conn.WriteMessage(websocket.TextMessage, []byte("Connection is established!"))
	if err != nil {
		log.Error().Msgf("Websocket send message error: %s", err)
		return
	}
}
