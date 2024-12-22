import nats from "nats";
import type { Context } from "./context.ts";

export const ping = (msg: nats.Msg) => {
    msg.respond();
};

export const send = (ctx: Context, msg: nats.Msg) => {
    msg.respond();
};
