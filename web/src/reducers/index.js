import { Reducers } from '../constants/reducers';
import { RequestRoute } from '../api/route';

const initial = {
    origin: null,
    destination: null,
    distance: 0.0,

    stops: [],

    ui: {
        updatingRoute: false,
        autocompleting: false,
    },
};

export default (state = initial, action) => {
    let next = Object.assign({}, state);

    switch (action.type) {
        /**
         * Make a request but don't actually change the state. The state will change when the 
         * request completes successfully.
         */
        case Reducers.ROUTE_ADD_REQUEST:
            let visit = next.stops.filter(stop => stop.RequestedStop).map(stop => stop.System.id)
            if (visit.indexOf(action.target) === -1) visit.push(action.target);

            RequestRoute(visit[0], null, visit.slice(1));

            // UI updates
            next.ui.updatingRoute = true;
            return next;

        case Reducers.ROUTE_UPDATE_SUCCESS:
            let route = action.route;

            // If stops, origin, and destination are specified, copy into state.
            if (route.Stops) next.stops = route.Stops;
            else if (route.Origin) next.stops = [route.Origin];
            else if (route.Destination) next.stops = [route.Destination];

            // Create collapsed legs.
            let collapsed = [];
            let complete = [];
            console.log('done');
            for (let i = 0; i < next.stops.length; i++) {
                if (!next.stops[i].RequestedStop) { // if can be collapsed
                    collapsed.push(next.stops[i]);
                } else {
                    if (collapsed.length > 0) { // if end of collapsed streak
                        complete.push({
                            Collapsed: true, 
                            TotalStops: collapsed.length, 
                            TotalDistance: collapsed.reduce((total, nx) => total + nx.DistanceFromPrev, 0),
                            Stops: collapsed,
                        });

                        collapsed = []; // reset collapsed
                    }

                    complete.push(next.stops[i]);
                }
            }
            console.log('done');
            next.stops = complete;

            // Origin, destination, and distance get aliased.
            if (route.Origin) next.origin = route.Origin.System;
            if (route.Destination) next.destination = route.Destination.System;
            if (route.Distance) next.distance = Math.floor(route.Distance * 100) / 100;

            console.log(`State:`, next);
            // UI updates
            next.ui.updatingRoute = false;
            return next;

        case Reducers.ROUTE_REMOVE_REQUEST:
            next.stops = next.stops
                .filter(stop => stop.RequestedStop)
                .filter(stop => stop.System.id !== action.target);
            
            RequestRoute(
                next.stops[0].System.id, 
                null, 
                next.stops.slice(1).map(stop => stop.System.id)
            );

            next.ui.updatingRoute = true;
            return next;

        default:
            return next;
    }
};