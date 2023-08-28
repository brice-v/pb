# pb - a very (very) small paste bin example

## Stuff

- No auth
- No rate limiting
- basic builtin UI
- simple api
- no config

Default URL is http://127.0.0.1:3001

### JSON enpoints

- POST /paste with `{'title': _, 'text': _}` JSON and it will return an ID int as JSON
- GET /paste/:id to retrieve the paste as JSON in format above

### UI

- / endpoint shows title and text form
- create new paste brings you to paste-ui view
- also access paste-ui view via /paste-ui/:id
- readonly in this view