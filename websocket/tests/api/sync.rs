use futures::{SinkExt, StreamExt};
use tokio_tungstenite::tungstenite;

use crate::helpers::spawn_app;

/* Generated via:
    ```js
       const doc = new Y.Doc()
       const ytext = doc.getText('type')
       doc..transact_mut()(function () {
           ytext.insert(0, 'def')
           ytext.insert(0, 'abc')
           ytext.insert(6, 'ghi')
           ytext.delete(2, 5)
       })
       const update = Y.encodeStateAsUpdate(doc)
       ytext.toString() // => 'abhi'
    ```
    a zero is appended for the sync message type
*/
const TEST_UPDATE: &[u8] = &[
    1, 5, 152, 234, 173, 126, 0, 1, 1, 4, 116, 121, 112, 101, 3, 68, 152, 234, 173, 126, 0, 2, 97,
    98, 193, 152, 234, 173, 126, 4, 152, 234, 173, 126, 0, 1, 129, 152, 234, 173, 126, 2, 1, 132,
    152, 234, 173, 126, 6, 2, 104, 105, 1, 152, 234, 173, 126, 2, 0, 3, 5, 2, 0,
];

#[tokio::test]
async fn other_client_receives_sync() {
    let app = spawn_app().await;
    let mut client_a = app.create_owner_client().await;
    let mut client_b = app.create_owner_client().await;
    let sync_payload: Vec<u8> = vec![1, 2, 3, 0];

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

#[tokio::test]
async fn get_diff_after_update() {
    let app = spawn_app().await;
    let mut client_a = app.create_owner_client().await;

    let sync_payload = TEST_UPDATE.to_vec();
    let get_diff_payload: Vec<u8> = vec![0, 1];

    client_a
        .send(tungstenite::Message::Binary(sync_payload.clone()))
        .await
        .unwrap();

    client_a
        .send(tungstenite::Message::Binary(get_diff_payload))
        .await
        .unwrap();

    let received = match client_a.next().await.unwrap().unwrap() {
        tungstenite::Message::Binary(payload) => payload,
        other => panic!("expected a binary message but got {other:?}"),
    };

    assert_eq!(sync_payload, received);
}
