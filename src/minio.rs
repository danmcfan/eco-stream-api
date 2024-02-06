use s3::{creds::Credentials, Bucket, Region};

pub async fn upload_file(path: &str, content: &[u8]) {
    let bucket = create_bucket_connection();
    bucket
        .put_object(path, content)
        .await
        .expect("Failed to upload file");
}

pub async fn download_file(path: &str) -> Vec<u8> {
    let bucket = create_bucket_connection();
    bucket
        .get_object(path)
        .await
        .expect("Failed to download file")
        .to_vec()
}

fn create_bucket_connection() -> Bucket {
    let bucket_name = std::env::var("MINIO_BUCKET").unwrap_or("default".to_string());
    let endpoint = std::env::var("MINIO_URL").unwrap_or("http://127.0.0.1:9000".to_string());

    let username = std::env::var("MINIO_ROOT_USER").unwrap_or("minioadmin".to_string());
    let password = std::env::var("MINIO_ROOT_PASSWORD").unwrap_or("minioadmin".to_string());

    Bucket::new(
        &bucket_name,
        Region::Custom {
            region: "us-east-1".to_string(),
            endpoint,
        },
        Credentials::new(Some(&username), Some(&password), None, None, None)
            .expect("Failed to create Minio credentials"),
    )
    .expect("Failed to create bucket connection")
    .with_path_style()
}
