# messaging-app-api

This is the backend for a messaging app called "StarSend" that allows users to register accounts for privately messaging others. Users are also able to customize their profiles, which show up when being searched by others. The frontend is built with React + TypeScript and is styled with Tailwind. The backend is built with Go and interacts with a PostgreSQL database.

The backend is main focus of this project, so there's much more that I experimented with and learned from making it!

The frontend repository can be found here: https://github.com/ken-ux/messaging-app

## File Structure

These are the folders in this project and information about the files contained in them.

- api
  - `auth.go`: Contains authentication functions including JWT generation and parsing.
  - `hash.go`: Password hashing utility functions.
  - `message.go`: Sending new messages and querying past messages.
  - `profile.go`: Getting, deleting, and updating profile settings.
  - `search.go`: Searching for users in database.
- db
  - `db.go`: Initializes connection to database.
- defs
  - `defs.go`: Type definitions.
- ws
  - `hub.go`: Initializes WebSocket connections then listens/responds to incoming or outgoing messages.

## Lessons Learned

- Generating a JWT with relevant claims encoded, such as the sub and expiration time.
  - Part of this process included getting a string (the secret) from the .env file and encoding it into bytes using Go's standard library. The Go library used for JWTs required the secret to be in bytes.
- Setting up CORS policy to include only specific headers and origins.
  - Configuring WebSocket connections to also check through the list of allowed origins before opening connections.
- Something I had to consider was whether I wanted to use the WebSocket API to handle messages sent between users rather than a standard REST API with POST requests. I went with the latter for the following reasons:
  - It's overkill to set-up WebSockets to authenticate a request, store the data in the database, then relay a message back to the client when these are inherent in regular REST API development.
  - Websockets is more suited for real-time notifications when a new message is received rather than handling the messaging process itself.
- In my type definitions, I add an "omitempty" tag to fields so that they are omitted during JSON marshalling when empty. However, this wasn't working when the field was `time.Time` because that field itself is a struct. Changing it to a pointer (`*time.Time`) fixes the issue.
  - This is because nil pointers are treated as empty, whereas the "zero" value of `time.Time` is still treated as valid and therefore doesn't got omitted.
- Passing a slice by a pointer in the `queryMessages` method of the `message.go` file so that the underlying slice is mutated instead of a shallow copy.
- I was using a LEFT JOIN instead of INNER JOIN in my SQL query for messages between two users, which prevented the query from filtering out records that contain non-valid recipients.
- Upserting a record in the database i.e. updating an existing row if a specified value already exists in a table, and inserting a new row if the specified value doesn't already exist.
- Using UNIQUE restraint on the `settings` table of the database so that it has a one-to-one relationship with the `users` table, ensuring that each user has only one row storing their settings data.
- Added ON DELETE CASCADE option to table keys so that messages associated with a user are deleted if the user is deleted from the database.

## To-Do

This is still a work-in-progress. The next step for this backend is the following:

- Developing WebSocket notifications around an event-driven architecture.
  - First, a message is sent through the WebSocket connection when a user makes a message.
  - Other users on an open WebSocket connection are polled to see if they are the relevant party in the message.
  - If they are, then we send the user a notification.
