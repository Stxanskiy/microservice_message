CREATE TABLE chats (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(255),
                       is_group BOOLEAN NOT NULL DEFAULT false,
                       created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE chat_participants (
                                   chat_id INT REFERENCES chats(id) ON DELETE CASCADE,
                                   user_id INT REFERENCES users(id) ON DELETE CASCADE,
                                   joined_at TIMESTAMP NOT NULL DEFAULT now(),
                                   PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE messages (
                          id SERIAL PRIMARY KEY,
                          chat_id INT REFERENCES chats(id) ON DELETE CASCADE,
                          sender_id INT REFERENCES users(id),
                          content TEXT,
                          created_at TIMESTAMP NOT NULL DEFAULT now()
);
