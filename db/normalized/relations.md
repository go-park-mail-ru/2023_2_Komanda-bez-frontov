# Таблицы

## User — Таблица с данными о пользователе

| Название | Тип |
|----|----|
| *id* | *integer* |
| first_name | varchar |
| last_name | varchar |
| email | varchar |
| username | varchar |
| password | varchar |

#### Функциональные зависимости
{ id } -> first_name, last_name, email, username, password

{ email } -> first_name, last_name, username, password

#### Нормальные формы:
  + **1 НФ** - поля id, first_name, last_name, email, username, password являются атомарными
  
  + **2 НФ** -  first_name, last_name, password функционально зависят полностью от первичных ключей id, email, составных ключей нет
  
  + **3 НФ** - среди неключевых атрибутов  first_name, last_name, username, password нет функциональных зависимостей
  
  + **НФБК** - Отношение находится в 3НФ и мы не имеем составного ключа 
    
## Form — Таблица с данными о существующих опросах

| Название  | Тип |
|-----------|-----|
| *id*        | *integer* |
| created_at |  time |
| author_id | integer |
| title | varchar |
	
#### Функциональные зависимости

{ id } -> created_at, author_id, title

#### Нормальные формы:
  + **1 НФ** - поля id, created_at, author_id, title являются атомарными
  
  + **2 НФ** -  created_at, author_id, title функционально зависят полностью от первичного ключа id, составных ключей нет
  
  + **3 НФ** - среди неключевых атрибутов created_at, author_id, title нет функциональных зависимостей
  
  + **НФБК** - Отношение находится в 3НФ и мы не имеем составного ключа 

## Question — Таблица с данными о составленных вопросах

| Название | Тип |
|----|----|
| *id* | *integer* |
| form_id | integer |
| question_type | varchar |
| question_title | string |
| question_text | text |
| shuffle | boolean |

#### Функциональные зависимости

{ id } -> form_id, question_type, question_title, question_text, shuffle

#### Нормальные формы:
  + **1 НФ** - поля id, form_id, question_type, question_title, question_text, shuffle являются атомарными
  
  + **2 НФ** -  orm_id, question_type, question_title, question_text, shuffle функционально зависят полностью от первичного ключа id, составных ключей нет
  
  + **3 НФ** - среди неключевых атрибутов orm_id, question_type, question_title, question_text, shuffle нет функциональных зависимостей
  
  + **НФБК** - Отношение находится в 3НФ и мы не имеем составного ключа
   
## Answer— Таблица с данными об предлагаемых ответах на вопрос

| Название | Тип |
|----|----|
| *id* | *integer* |
| question_id | integer |
| answer_text | text |

#### Функциональные зависимости

{ id } -> question_id, answer_text

#### Нормальные формы:
  + **1 НФ** - поля id, question_id, answer_text являются атомарными
  
  + **2 НФ** -  question_id, answer_text функционально зависят полностью от первичного ключа id, составных ключей нет
  
  + **3 НФ** - среди неключевых атрибутов question_id, answer_text нет функциональных зависимостей
  
  + **НФБК** - Отношение находится в 3НФ и мы не имеем составного ключа

## Form_passage — Таблица с данными о прохождении опроса

| Название | Тип |
|----|----|
| *id* | *integer* |
| user_id | integer |
| form_id | integer |
| started_at | time |

#### Функциональные зависимости

{ id } -> user_id, form_id, started_at

#### Нормальные формы:
  + **1 НФ** - поля id, user_id, form_id, started_at являются атомарными
  
  + **2 НФ** -  user_id, form_id, started_at функционально зависят полностью от первичного ключа id, составных ключей нет
  
  + **3 НФ** - среди неключевых атрибутов user_id, form_id, started_at нет функциональных зависимостей
  
  + **НФБК** - Отношение находится в 3НФ и мы не имеем составного ключа

## Form_passage_answer — Таблица с данными об ответах на пройденный опрос

| Название | Тип |
|----|----|
| *id* | *integer* |
|form_passage_id | integer |
| question_id | integer |
| answer_text | text |

#### Функциональные зависимости

{ id } -> form_passage_id, question_id, answer_text

#### Нормальные формы:
  + **1 НФ** - поля id, form_passage_id, question_id, answer_text являются атомарными
  
  + **2 НФ** -  form_passage_id, question_id, answer_text функционально зависят полностью от первичного ключа id, составных ключей нет
  
  + **3 НФ** - среди неключевых атрибутов form_passage_id, question_id, answer_text нет функциональных зависимостей
  
  + **НФБК** - Отношение находится в 3НФ и мы не имеем составного ключа

# ER Diagram

Диаграмма БД представлена ниже.

[А так же ссылка на веб версию](https://erd.dbdesigner.net/designer/schema/1698230688-formhub)

![image](https://github.com/go-park-mail-ru/2023_2_Komanda-bez-frontov/assets/114286666/d50fb24f-bac4-49eb-9bf1-188a7a14a68b)

```
User {
	id integer pk increments unique
	first_name varchar
	last_name varchar
	email varchar unique
	username varchar unique
	password varchar
}

Form {
	id integer pk increments unique
	created_at time
	author_id integer *> User.id
	title varchar
}

Question {
	id integer pk increments unique
	form_id integer *> Form.id
	question_type varchar
	question_title string
	question_text text
	shuffle boolean
}

Answer {
	id integer pk increments
	question_id integer *> Question.id
	answer_text text
}

Form_passage {
	id integer pk increments
	user_id integer *> User.id
	form_id integer *> Form.id
	started_at time
}

Form_passage_answer {
	id integer pk increments unique
	form_passage_id integer *> Form_passage.id
	question_id integer > Question.id
	answer_text text
}
```
