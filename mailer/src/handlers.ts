import nats from "nats";
import * as srv from "./service.ts"
import type { Context } from "./context.ts";

export const ping = (msg: nats.Msg) => {
    msg.respond(JSON.stringify({
        message: "pong"
    }))
};

interface ISentInput {
    receiver: string,
    templateName: string,
    options: Map<string, unknown>
}

export const send = (ctx: Context, msg: nats.Msg) => {
    const req = msg.json<ISentInput>()

};
