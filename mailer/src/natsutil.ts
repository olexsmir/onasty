import nats from "nats";
import type { Context } from "./context.ts";

export type SubscriberCallbackFunction = (msg: nats.Msg) => void;

export async function subscribe(
    conn: nats.NatsConnection,
    subject: string,
    callback: SubscriberCallbackFunction,
) {
    const s = conn.subscribe(subject);
    for await (const m of s) {
        callback(m);
    }
}

export type SubscriberCallbackWithContextFunction = (
    ctx: Context,
    msg: nats.Msg,
) => void;

export async function subscribeContext(
    conn: nats.NatsConnection,
    subject: string,
    ctx: Context,
    callback: SubscriberCallbackWithContextFunction,
) {
    const s = conn.subscribe(subject);
    for await (const m of s) {
        callback(ctx, m);
    }
}
