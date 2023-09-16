CREATE TABLE user_login (
    user_id VARCHAR(255) PRIMARY KEY,
    user_username VARCHAR(255) NOT NULL,
    bycrypt_password VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE user_information (
    user_id VARCHAR(255) PRIMARY KEY,
    user_firstname VARCHAR(100) NOT NULL,
    user_lastname VARCHAR(100) NOT NULL,
    user_email VARCHAR(100) NOT NULL,
    user_channelID VARCHAR(30) NOT NULL,
)

CREATE TABLE channel_information (
    channnel_id SERIAL PRIMARY KEY,
    channel_name VARCHAR(100) NOT NULL,
    channel_dateCreated DATE NOT NULL,
    channel_subscriberCount INT NOT NULL,
    channel_videoCount INT NOT NULL,
)

-- Insert default user
--INSERT INTO users (sPassword, sUsername) VALUES ('$2a$10$gvAudYjysfI4zZTR.kRr5ezPV2qsYHCn.bzQtJ5ks4FSQYnb22gaq', 'user');