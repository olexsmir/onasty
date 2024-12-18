import nats from "nats"

export const send = (msg: nats.Msg) => {
    console.log(
        msg.json()
    );

}
