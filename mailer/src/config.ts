export default {
    appUrl: process.env.APP_URL || "http://localhost",
    natsUrl: process.env.NATS_URL || 'localhost:4222',
    mailgunApiKey: process.env.MAILGUN_API_KEY || "",
    mailgunDomain: process.env.MAILGUN_DOMAIN || "",
    mailgunFrom: process.env.MAILGUN_FROM || "",
}
