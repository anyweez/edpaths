import { Reducers } from '../constants/reducers';

export function addToRoute(id) {
    return {
        type: Reducers.ROUTE_ADD_REQUEST,
        target: id,
    };
}

export function updateRoute(route) {
    return {
        type: Reducers.ROUTE_UPDATE_SUCCESS,
        route: route,
    };
}

export function removeFromRoute(id) {
    return {
        type: Reducers.ROUTE_REMOVE_REQUEST,
        target: id,
    };
}