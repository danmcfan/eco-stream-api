use crate::models::User;
use redis::Commands;

pub fn list_users() -> redis::RedisResult<Vec<User>> {
    let mut conn = create_redis_connection();
    let users: Vec<String> = conn.hvals("users")?;

    Ok(users
        .iter()
        .map(|user| serde_json::from_str(user).unwrap())
        .collect())
}

pub fn retrieve_user(id: String) -> redis::RedisResult<Option<User>> {
    let mut conn = create_redis_connection();
    let user: Option<String> = conn.hget("users", id)?;

    match user {
        Some(user) => Ok(Some(serde_json::from_str(&user).unwrap())),
        None => Ok(None),
    }
}

pub fn store_user(user: &User) -> redis::RedisResult<()> {
    let mut conn = create_redis_connection();
    conn.hset("users", &user.id, serde_json::to_string(user).unwrap())?;
    Ok(())
}

pub fn delete_user(id: String) -> redis::RedisResult<()> {
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
