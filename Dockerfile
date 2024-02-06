FROM rust:1.75.0-slim-bullseye as builder

WORKDIR /usr/src/eco-stream

COPY . .

RUN apt-get update && apt-get install -y pkg-config libssl-dev && rm -rf /var/lib/apt/lists/*

RUN cargo install --path .

FROM debian:bullseye-slim


COPY --from=builder /usr/local/cargo/bin/eco-stream /usr/local/bin/eco-stream

EXPOSE 80

CMD ["eco-stream"]
