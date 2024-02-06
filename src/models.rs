use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
pub struct CreateUser {
    pub username: String,
}

#[derive(Deserialize, Serialize)]
pub struct User {
    pub id: String,
    pub username: String,
}
