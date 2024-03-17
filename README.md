**Example Documentation: Using Distcache with PostgreSQL**

**Overview:**

This example demonstrates how to integrate Distcache, a distributed caching library for Go, with a PostgreSQL database to improve application performance by caching data fetched from the database. The example showcases fetching and caching individual todo items as well as fetching and caching all todo items.

**Prerequisites:**

1. Go installed on your machine.
2. PostgreSQL database setup with a `todos` table.

**Setup:**

1. Clone the GitHub repository containing the example code:

   ```bash
   git clone <repository_url>
   ```

2. Navigate to the directory containing the example code:

   ```bash
   cd dist-cache-example
   ```

3. Install dependencies:

   ```bash
   go mod download
   ```

4. Start the PostgreSQL server and ensure it is running.

5. Run the provided SQL script (`schema.sql`) to create the `todos` table in the PostgreSQL database.

   ```bash
   psql -U <username> -d <database_name> -a -f schema.sql
   ```

**Usage:**

1. Start the application server by running the provided script:

   ```bash
   ./run.sh
   ```

   Alternatively, you can start the server manually by running:

   ```bash
   go run main.go
   ```

2. Once the server is running, you can use Postman or any other HTTP client to interact with the API endpoints.

**Endpoints:**

- `GET /api/todos?key={todo_id}`: Retrieves a todo item by its ID. If the item is not found in the cache, it is fetched from the database and cached.
- `GET /api/todos`: Retrieves all todo items. If all todos are requested for the first time, they are fetched from the database, cached, and returned. Subsequent requests for all todos retrieve them from the cache.
- `POST /api/todos`: Creates a new todo item in the database.

**Postman Collection:**

A Postman collection (`distcache-postgres-example.postman_collection.json`) is provided with pre-configured requests to test the API endpoints. Import the collection into Postman to easily test the endpoints.

**GitHub Repository:**

The source code for the Distcache:

[Distcache ](https://github.com/rohitlohar45/distcache)

**Conclusion:**

By integrating Distcache with a PostgreSQL database, this example demonstrates how to efficiently cache data fetched from the database, thereby reducing database load and improving application performance. Distcache provides a simple yet powerful caching solution for Go applications, offering flexibility and scalability for various caching needs.
