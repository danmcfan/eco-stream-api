mod auth;
mod handlers;
mod minio;
mod models;
mod redis;

use auth::auth;
use axum::{
    middleware,
    routing::{delete, get, post},
    Router,
};
use handlers::{
    create_user_handler, delete_user_handler, download_file_handler, health_handler,
    list_users_handler, retrieve_user_handler, upload_file_handler,
};

#[tokio::main]
async fn main() {
    let app = Router::new()
        .route("/health", get(health_handler))
        .route("/users", get(list_users_handler))
        .route("/users/:id", get(retrieve_user_handler))
        .route(
            "/users",
            post(create_user_handler).layer(middleware::from_fn(auth)),
        )
        .route(
            "/users/:id",
            delete(delete_user_handler).layer(middleware::from_fn(auth)),
        )
        .route("/download", get(download_file_handler))
        .route(
            "/upload",
            post(upload_file_handler).layer(middleware::from_fn(auth)),
        );

    let listener_url =
        std::env::var("LISTENER_URL").unwrap_or_else(|_| String::from("127.0.0.1:8080"));
    let listener = tokio::net::TcpListener::bind(listener_url).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
