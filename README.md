Order Service — README

What is this
Service for managing orders. Saves orders to database and calls Payment Service to process payments.

Architecture (Clean Architecture)

Domain — models and business rules

Use Case — business logic (create, cancel order)

Repository — all database queries

Transport — HTTP handlers, only parse request and return response

Each layer depends only inward. Handlers do not contain business logic.

Bounded Context

Order Service owns only orders and its own database. It does not touch Payment Service database. Communication only through HTTP.

Flow when creating an order

Save order with status Pending

Call Payment Service POST /payments

Get response → update status to Paid or Failed

Failure Handling

If Payment Service is down — HTTP client times out after 2 seconds, order is marked Failed, user gets 503. We chose Failed over Pending because if payment never happened the order clearly did not succeed.

Business Rules

Amount must be greater than 0

Only Pending orders can be cancelled

Paid orders cannot be cancelled


Payment Service — README

What is this

Service for processing payments. Receives requests from Order Service, checks amount limit, saves payment to its own database.

Architecture (Clean Architecture)

Same layer structure as Order Service. Each layer has one responsibility. Business logic lives only in Use Case and Domain.

Bounded Context

Payment Service owns only payments and its own database. Does not know about orders. Only receives order_id and amount.

Business Rules

Amount greater than 100000 → Declined

Amount less or equal to 100000 → Authorized

Each payment gets a unique transaction_id

Failure Handling

If Payment Service goes down — Order Service handles it with a 2 second timeout on its side.