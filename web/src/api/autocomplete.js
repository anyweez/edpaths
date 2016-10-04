export function RequestCompletion(fragment, cb) {
    fetch(`http://localhost:8080/search?q=${fragment}`)
        .then(res => res.json())
        .then(options => cb(options))
        .catch(err => {
            console.error(`API request error: ${err}`);
        });
}