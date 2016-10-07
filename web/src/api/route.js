import { updateRoute } from '../actions/route';
import store from '../store';

/**
 * Request a new route using the specified start and stop system ID's. This function 
 * expects the first two parameters to be integer ID's and the third to be an array of
 * integer ID's. These ID's should have been provided by the server previously.
 */
export function RequestRoute(start, stop, visit) {
    console.log('RequestRoute:')
    console.log(start, stop, visit);
    let stops = visit.join(',');
    fetch(`/route?from=${start}&visit=${stops}`)
        .then(res => res.json())
        .then(result => {
            console.log(result);
            store.dispatch(updateRoute(result.Route));
        })
        // TODO: no route found
        .catch(err => console.error(`API request error: ${err}`));
}