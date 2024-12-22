import type { IMailgunClient } from "mailgun.js/Interfaces"

export interface Context {
    mailgun: IMailgunClient,
}
