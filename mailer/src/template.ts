import config from "./config.ts"

export type Template = {
    subject: string,
    body: string,
}

export const emailVerification = (token: string): Template => {
    return {
        subject: "Onasty: verify your email",
        body: `To verify your email, please follow this link:
<a href="${config.appUrl}/api/v1/auth/verify/${token}">${config.appUrl}/api/v1/auth/verify/${token}</a>
<br />
<br />
This link will expire after 24 hours.`
    }
}
