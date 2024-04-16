use axum::{
    extract::{Request, State},
    http::{header, StatusCode},
    middleware::Next,
    response::Response,
};
use jsonwebtoken::{decode, DecodingKey, TokenData, Validation};
use secrecy::{ExposeSecret, Secret};
use serde::{Deserialize, Serialize};
use serde_aux::field_attributes::deserialize_number_from_string;

#[derive(Debug, Serialize, Deserialize)]
pub struct Claims {
    #[serde(deserialize_with = "deserialize_number_from_string")]
    exp: u64,
    user_id: String,
    username: String,
}

#[derive(Clone)]
pub struct User {
    id: String,
    username: String,
}

pub async fn auth_middleware(
    State(signing_key): State<Secret<String>>,
    mut req: Request,
    next: Next,
) -> Result<Response, StatusCode> {
    let auth_header = req
        .headers()
        .get(header::AUTHORIZATION)
        .and_then(|header| header.to_str().ok());

    let auth_header = if let Some(auth_header) = auth_header {
        auth_header
    } else {
        return Err(StatusCode::UNAUTHORIZED);
    };

    let mut auth_header_parts = auth_header.split(" ");
    if auth_header_parts.next() != Some("Bearer") {
        return Err(StatusCode::UNAUTHORIZED);
    };

    let token_string = if let Some(token) = auth_header_parts.next() {
        token
    } else {
        return Err(StatusCode::UNAUTHORIZED);
    };

    let token = if let Ok(token) = decode_jwt(token_string, signing_key) {
        token
    } else {
        // TODO: Handle different kind of decoding errors
        return Err(StatusCode::UNAUTHORIZED);
    };

    let user = User {
        id: token.claims.user_id,
        username: token.claims.username,
    };
    req.extensions_mut().insert(user);

    Ok(next.run(req).await)
}

fn decode_jwt(
    token: &str,
    signing_key: Secret<String>,
) -> jsonwebtoken::errors::Result<TokenData<Claims>> {
    decode(
        &token,
        &DecodingKey::from_secret(signing_key.expose_secret().as_ref()),
        &Validation::new(jsonwebtoken::Algorithm::HS256),
    )
}
