const SmeeClient = require('smee-client')

console.log(process.env.SMEE)
const smee = new SmeeClient({
    source: process.env.SMEE,
    target: 'http://localhost:8080/api/v1/hook/github',
    logger: console
})

const events = smee.start()

// Stop forwarding events
// events.close()