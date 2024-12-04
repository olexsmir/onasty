import nats from "nats"

export type SubscriberCallbackFunction = (msg: nats.Msg) => void

export async function subscribe(
    conn: nats.NatsConnection,
    subject: string,
    callback: SubscriberCallbackFunction
) {
    const s = conn.subscribe(subject)
    for await (const m of s) {
        callback(m)
    }
}
