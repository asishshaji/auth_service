-- DROP TYPE IF EXISTS user_status CASCADE;

-- CREATE TYPE user_status AS ENUM('enabled', 'disabled', 'blacklisted');

-- DROP TABLE IF EXISTS users CASCADE;

-- CREATE TABLE users (
  --  username varchar(40) PRIMARY KEY,
    -- password varchar(120) NOT NULL,
   -- company varchar(30) NOT NULL,
   -- status user_status NOT NULL DEFAULT 'enabled',
   -- created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
   -- updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
-- );

DROP INDEX IF EXISTS idx_user_company;

CREATE INDEX idx_user_company ON users(company);
