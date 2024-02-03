use axum::{
    extract::Path,
    http::StatusCode,
    response::IntoResponse,
    routing::{delete, get, post},
    Json, Router,
};
use redis::Commands;
use serde::{Deserialize, Serialize};
use serde_json::json;
use uuid::Uuid;

#[derive(Deserialize)]
struct CreateUser {
    username: String,
}

#[derive(Deserialize, Serialize)]
struct User {
    id: String,
    username: String,
}

#[tokio::main]
async fn main() {
    let app = Router::new()
        .route("/health", get(health_handler))
        .route("/users", get(list_users_handler))
        .route("/users/:id", get(retrieve_user_handler))
        .route("/users", post(create_user_handler))
        .route("/users/:id", delete(delete_user_handler));

    let listener_url =
        std::env::var("LISTENER_URL").unwrap_or_else(|_| String::from("127.0.0.1:8080"));
    let listener = tokio::net::TcpListener::bind(listener_url).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn health_handler() -> impl IntoResponse {
    (StatusCode::OK, "All systems operational!")
}

async fn list_users_handler() -> impl IntoResponse {
    match list_users() {
        Ok(users) => (StatusCode::OK, Json(json!(users))),
        Err(_) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": "Failed to list users"})),
        ),
    }
}

async fn retrieve_user_handler(Path(id): Path<String>) -> impl IntoResponse {
    match retrieve_user(id) {
        Ok(Some(user)) => (StatusCode::OK, Json(json!(user))),
        Ok(None) => (
            StatusCode::NOT_FOUND,
            Json(json!({"error": "User not found"})),
        ),
        Err(_) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": "Failed to retrieve user"})),
        ),
    }
}

async fn create_user_handler(Json(payload): Json<CreateUser>) -> impl IntoResponse {
    let user = User {
        id: Uuid::new_v4().to_string(),
        username: payload.username,
    };

    store_user(&user).expect("Failed to store user");

    (StatusCode::CREATED, Json(user))
}

async fn delete_user_handler(Path(id): Path<String>) -> impl IntoResponse {
    match delete_user(id) {
        Ok(_) => (StatusCode::NO_CONTENT, Json(json!(""))),
        Err(_) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": "Failed to delete user"})),
        ),
    }
}

fn list_users() -> redis::RedisResult<Vec<User>> {
    let mut conn = create_redis_connection();
    let users: Vec<String> = conn.hvals("users")?;

    Ok(users
        .iter()
        .map(|user| serde_json::from_str(user).unwrap())
        .collect())
}

fn retrieve_user(id: String) -> redis::RedisResult<Option<User>> {
    let mut conn = create_redis_connection();
    let user: Option<String> = conn.hget("users", id)?;

    match user {
        Some(user) => Ok(Some(serde_json::from_str(&user).unwrap())),
        None => Ok(None),
    }
}

fn store_user(user: &User) -> redis::RedisResult<()> {
    let mut conn = create_redis_connection();
    conn.hset("users", &user.id, serde_json::to_string(user).unwrap())?;
    Ok(())
}

fn delete_user(id: String) -> redis::RedisResult<()> {
    let mut conn = create_redis_connection();
    conn.hdel("users", id)?;
    Ok(())
}

fn create_redis_connection() -> redis::Connection {
    let redis_url =
        std::env::var("REDIS_URL").unwrap_or_else(|_| String::from("redis://127.0.0.1/"));
    let client = redis::Client::open(redis_url).expect("Failed to create Redis client");
    client
        .get_connection()
        .expect("Failed to get Redis connection")
}
