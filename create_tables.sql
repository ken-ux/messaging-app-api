-- Create user table.
CREATE TABLE
    "user" ( -- This is in double quotes because user is a reserved word.
        user_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        username VARCHAR(20) NOT NULL,
        password VARCHAR(72) NOT NULL
    );

-- Create table for managing user settings and preferences.
CREATE TABLE
    settings (
        settings_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        user_id INTEGER NOT NULL,
        description VARCHAR(100),
        color BYTEA,
        FOREIGN KEY (user_id) REFERENCES "user" (user_id) -- "user" is in double quotes it's a reserved word.
    );

-- Create table for message data.
CREATE TABLE
    message (
        message_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        author_id INTEGER NOT NULL,
        message_body VARCHAR(200) NOT NULL,
        creation_date TIMESTAMP NOT NULL,
        FOREIGN KEY (author_id) REFERENCES "user" (user_id)
    );

-- Create table to link recipients to messages.
CREATE TABLE
    recipient (
        recipient_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        user_id INTEGER NOT NULL,
        message_id INTEGER NOT NULL,
        FOREIGN KEY (user_id) REFERENCES "user" (user_id),
        FOREIGN KEY (message_id) REFERENCES message (message_id)
    );