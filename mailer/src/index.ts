import nats from "nats"
import * as natsutil from "./natsutil.ts"
import config from "./config.ts"
import * as handlers from "./handlers.ts"
import Mailgun from "mailgun.js"
import formData from 'form-data'

(async () => {
    const nc = await nats.connect({ servers: config.natsUrl });
    const mailgun = new Mailgun(formData).client({
        username: 'api',
        key: config.mailgunApiKey,

    });


    natsutil.subscribe(nc, "mailer.send", handlers.send);
})()
