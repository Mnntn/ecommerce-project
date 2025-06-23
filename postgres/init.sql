CREATE DATABASE orders;
CREATE DATABASE payments;

\c orders
\i /app/migrations/order-service/001_init.sql

\c payments
\i /app/migrations/payment-service/001_init.sql 