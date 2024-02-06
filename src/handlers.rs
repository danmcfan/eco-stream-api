use crate::minio::{download_file, upload_file};
use crate::models::{CreateUser, User};
use crate::redis::{delete_user, list_users, retrieve_user, store_user};
use axum::{extract::Path, http::StatusCode, response::IntoResponse, Json};
use serde_json::json;
use uuid::Uuid;

pub async fn health_handler() -> impl IntoResponse {
    (StatusCode::OK, "All systems operational!")
}

pub async fn list_users_handler() -> impl IntoResponse {
    match list_users() {
        Ok(users) => (StatusCode::OK, Json(json!(users))),
        Err(_) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": "Failed to list users"})),
        ),
    }
}

pub async fn retrieve_user_handler(Path(id): Path<String>) -> impl IntoResponse {
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

pub async fn create_user_handler(Json(payload): Json<CreateUser>) -> impl IntoResponse {
    let user = User {
        id: Uuid::new_v4().to_string(),
        username: payload.username,
    };

    store_user(&user).expect("Failed to store user");

    (StatusCode::CREATED, Json(user))
}

pub async fn delete_user_handler(Path(id): Path<String>) -> impl IntoResponse {
    match delete_user(id) {
        Ok(_) => (StatusCode::NO_CONTENT, Json(json!(""))),
        Err(_) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": "Failed to delete user"})),
        ),
    }
}

pub async fn download_file_handler() -> impl IntoResponse {
    let content = download_file("/hello.txt").await;
    (StatusCode::OK, content)
}

pub async fn upload_file_handler() -> impl IntoResponse {
    let content = b"Hello, world!\n";
    upload_file("/hello.txt", content).await;
    (StatusCode::CREATED, "File uploaded!")
}
