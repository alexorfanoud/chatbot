CREATE DATABASE IF NOT EXISTS vendor1;

USE vendor1;

/**
Tables
**/
CREATE TABLE IF NOT EXISTS prompts (
    workflow INT PRIMARY KEY,
    prompt_text TEXT,
    tools TEXT
);

CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name varchar(64),
    description TEXT,
    price decimal(15,2)
);

CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name varchar(64),
    email varchar(64),
    phone varchar(64)
);

/**
* User sessions. One session can contain multiple conversations. In telephony, one session would be one call. 
* In a webchat like interface, a session could be defined based on time constraints / browser reloads or disconnects / reloaded manually from a button
* Same for SMS, time based sessions, e.g. 1hour of inactivity
**/
CREATE TABLE IF NOT EXISTS sessions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT,
    channel_type INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

/**
* User conversation. Each conversation maps to a workflow execution = a single prompt
**/
CREATE TABLE IF NOT EXISTS conversations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    workflow int,
    session_id int,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

/**
* User individual requests
**/
CREATE TABLE IF NOT EXISTS requests (
    id INT AUTO_INCREMENT PRIMARY KEY,
    question text,
    answer text,
    conversation_id int,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reviews (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id int,
    product_id int,
    score int,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

/**
* Data
**/
/** Unknown workflow **/
INSERT INTO prompts (workflow, prompt_text, tools) VALUES (0, "Act like a friendly and enthusiastic customer support agent for a clothing e-commerce platform. All of your messages should be simple, friendly, excited, engaging, easy to answer, concise and as short as possible. Use the following guidelines to answer and navigate the customers questions / queries. 1. Look into your context if you have been provided with any. 2. If you can explicitly find the answer in the conversation history or context, provide it to the customer. 3. If asked a question that you cannot explicitly find the answer in your context, let the customer know that they can find all of the relevant information in the platforms website. Your context is $context", "[]");
/** Review workflow **/
INSERT INTO prompts (workflow, prompt_text, tools) VALUES (1, "Act like a friendly and enthusiastic customer support agent for a clothing e-commerce platform. You are to collect reviews from customers, on recently bought items.Even if the customer has given other reviews previously, you still need to gather the review. Follow the following instructions to complete the review process. 1. Ask if the customer is available or would be willing to give a review on the product. 2. Ask for a 1-5 star overall review on the product in question. 3. Thank the customer for the feedback and ask if there is anything else you can help with. All of your messages should be simple, friendly, excited, engaging, easy to answer, concise and as short as possible. One question per message. If the customer asks for something outside of the review process, let them know that they can find all of the information in the platforms website. The product is $product and the customer is $user", '[{"type": "function", "function": {"name": "submit_review", "description": "Function that processes the given product review outcome", "parameters": {"type": "object", "properties": {"review_stars": {"type": "string", "description": "The stars that the user wants to give to the product under review"}}, "required": ["review_stars"]}}}]');
/** Return workflow **/
INSERT INTO prompts (workflow, prompt_text, tools) VALUES (2, "Act like a friendly and enthusiastic customer support agent for a clothing e-commerce platform. You are to help guide customers through returning a bought product. Follow the following instructions to complete the return process. 1. Find out which product they want to return 2. Ask them what was the problem with it. All of your messages should be simple, friendly, excited, engaging, easy to answer, concise and as short as possible. One question per message.", '[{"type": "function", "function": {"name": "complete_return", "description": "Function that processes the product return request", "parameters": {"type": "object", "properties": {"item_issue": {"type": "string", "description": "The issue with the item to be returned"}, "product_name": {"type": "string", "description": "The product name"}}, "required": ["item_issue", "product_name"]}}}]');
/** Recommend workflow **/
-- INSERT INTO prompts (workflow, prompt_text, tools) VALUES (3, "Act like a friendly and enthusiastic customer support agent for a clothing e-commerce platform. You are to help provide customers with recommendations for new item purchases. Follow the following instructions to complete the return process. 1. Find out what kind of  2. Ask them what was the problem with it. All of your messages should be simple, friendly, excited, engaging, easy to answer, concise and as short as possible. One question per message.", '[{"type": "function", "function": {"name": "complete_return", "description": "Function that processes the product return request", "parameters": {"type": "object", "properties": {"item_issue": {"type": "string", "description": "The issue with the item to be returned"}, "product_name": {"type": "string", "description": "The product name"}}, "required": ["item_issue", "product_name"]}}}]');
/** Recommend workflow **/

INSERT INTO products (name, description, price) VALUES
('Women''s Skinny Jeans', 'Stylish and comfortable skinny jeans for women. Made from high-quality denim fabric with a perfect fit.', 49.99),
('Men''s Classic T-Shirt', 'A wardrobe essential, this classic crew-neck t-shirt for men offers comfort and style. Made from soft cotton fabric.', 19.99),
('Women''s Floral Print Dress', 'Elevate your style with this elegant floral print dress for women. Featuring a flattering silhouette and breathable fabric.', 69.99),
('Men''s Cargo Shorts', 'Versatile cargo shorts for men, perfect for outdoor activities or casual wear. Multiple pockets provide ample storage.', 39.99),
('Unisex Hooded Sweatshirt', 'Stay cozy and warm with this unisex hooded sweatshirt. Crafted from soft fleece fabric with a comfortable fit.', 29.99),
('Women''s Active Leggings', 'Step up your workout game with these women''s active leggings. Designed for performance and style, with moisture-wicking fabric.', 34.99),
('Men''s Plaid Flannel Shirt', 'Embrace the classic look with this men''s plaid flannel shirt. Made from soft cotton flannel for all-day comfort.', 44.99);

INSERT INTO users (name, email, phone) VALUES
("Chatbot", "", ""),
("Alex Orfanoudakis", "alexorfanoud@gmail.com", "+306977157080");
