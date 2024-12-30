import type { IMailgunClient } from "mailgun.js/Interfaces";

export interface Context {
    mailgun: IMailgunClient
}

export interface ISentInput {
    receiver: string,
    templateName: string,
    options: Map<string, unknown>
}
