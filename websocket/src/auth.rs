use std::str::FromStr;

use axum::{
    extract::{Request, State},
    http::header,
    middleware::Next,
    response::Response,
};
use jsonwebtoken::{decode, DecodingKey, TokenData, Validation};
use secrecy::{ExposeSecret, Secret};
use serde::{Deserialize, Serialize};
use serde_aux::field_attributes::deserialize_number_from_string;
use uuid::Uuid;

use crate::error::ApiError;

#[derive(Debug, Serialize, Deserialize)]
pub struct Claims {
    #[serde(deserialize_with = "deserialize_number_from_string")]
    pub exp: u64,
    pub user_id: String,
    pub username: String,
}

#[derive(Debug, Clone)]
pub struct User {
    pub id: Uuid,
    pub username: String,
}

pub async fn auth_middleware(
    State(signing_key): State<Secret<String>>,
    mut req: Request,
    next: Next,
) -> Result<Response, ApiError> {
    let auth_header = req
        .headers()
        .get(header::AUTHORIZATION)
        .and_then(|header| header.to_str().ok())
        .ok_or_else(|| ApiError::AuthError("auth header is missing".to_string()))?;

    let mut auth_header_parts = auth_header.split(" ");
    if auth_header_parts.next() != Some("Bearer") {
        return Err(ApiError::AuthError(
            "auth header bearer prefix missing".to_string(),
        ));
    };

    let token_string = auth_header_parts
        .next()
        .ok_or_else(|| ApiError::AuthError("bearer token missing".to_string()))?;

    let token = decode_jwt(token_string, signing_key).map_err(|e| {
        tracing::error!(?e, "JWT decoding error");
        ApiError::AuthError("invalid token".to_string())
    })?;

    let user = User {
        id: Uuid::from_str(&token.claims.user_id).unwrap(),
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
