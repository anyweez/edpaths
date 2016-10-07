export function RequestCompletion(fragment, cb) {
    fetch(`/search?q=${fragment}`)
        .then(res => res.json())
        .then(options => cb(options))
        .catch(err => {
            console.error(`API request error: ${err}`);
        });
}