CREATE TABLE users (
    id INTEGER NOT NULL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    salt TEXT NOT NULL,
    created_at DATETIME DEFAULT (datetime('now')) NOT NULL
);

CREATE TABLE wallets (
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT (datetime('now')) NOT NULL
);

CREATE TABLE wallet_users (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id INTEGER NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, wallet_id)
);

CREATE TABLE tags (
    id INTEGER NOT NULL PRIMARY KEY,
    wallet_id INTEGER NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT (datetime('now')) NOT NULL,
    UNIQUE(wallet_id, name)
);

CREATE TABLE transactions (
    id INTEGER NOT NULL PRIMARY KEY,
    wallet_id INTEGER NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    value INTEGER NOT NULL,
    tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT (datetime('now')) NOT NULL
);
