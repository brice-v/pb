package sql

const CreatePasteTable = `CREATE TABLE IF NOT EXISTS pastes(id INTEGER PRIMARY KEY, title TEXT, paste_text TEXT NOT NULL);`
const GetMaxPasteId = `SELECT MAX(id) FROM pastes;`
const InsertPasteTable = `INSERT INTO pastes (id, title, paste_text) VALUES (?, ?, ?);`
const GetPaste = `SELECT title, paste_text FROM pastes WHERE id = ?;`
