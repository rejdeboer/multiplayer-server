use futures::{SinkExt, StreamExt};
use tokio_tungstenite::tungstenite;

use crate::helpers::spawn_app;

#[tokio::test]
async fn ping_pong_works() {
    let app = spawn_app().await;
    let mut client_a = app.create_owner_client().await;
    let mut client_b = app.create_owner_client().await;
    let sync_payload: Vec<u8> = vec![0, 1, 2, 3];

    client_a
        .send(tungstenite::Message::Binary(sync_payload.clone()))
        .await
        .unwrap();

    let received = match client_b.next().await.unwrap().unwrap() {
        tungstenite::Message::Binary(payload) => payload,
        other => panic!("expected a binary message but got {other:?}"),
    };

    assert_eq!(sync_payload, received);
}
