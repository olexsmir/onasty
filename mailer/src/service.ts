import config from "./config.ts";
import type { Context, ISentInput } from "./context";
import * as tmpl from "./template.ts"

export const send = (
    ctx: Context,
    templateName: string,
    receiver: string,
    options: <string, string>
) => {
    if (templateName === "email_verification") {
        const template = tmpl.emailVerification(options["token"])


        ctx.mailgun.messages.create(config.mailgunDomain, {

        })

    }



}
