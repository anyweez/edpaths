import React, { Component } from 'react';

import { RouteStop } from './RouteStop';
import { RouteLeg } from './RouteLeg';

export class RouteList extends Component {
    render() {
        // Generate all stops
        let stops = this.props.stops.map((stop, i) => {
            if (stop.Collapsed) return (<RouteLeg key={i} leg={stop}></RouteLeg>);
            else return (<RouteStop key={i} system={stop}></RouteStop>);
        });

        return (
            <ul className="route-list">
                {stops}
            </ul>
        );
    }
}