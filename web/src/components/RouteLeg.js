import React, { Component } from 'react';
import { RouteList } from './RouteList';

export class RouteLeg extends Component {
    constructor() {
        super();
        this.state = { hidden: true };
    }

    reveal() {
        this.setState({ hidden: !this.state.hidden }, () => console.log(this.state.hidden));
    }

    render() {
        let leg = this.props.leg;
        console.log('class: ', this.state.hidden ? "hidden" : "")

        return (
            <li className="route-leg">
                <h2 onClick={() => this.reveal()}>{Math.round(leg.TotalDistance * 100) / 100} ly, {leg.TotalStops} stops</h2>
                <div className={this.state.hidden ? "hide" : ""}>
                    <RouteList stops={leg.Stops}></RouteList>
                </div>
            </li>
        );
    }
}