import React, { Component } from 'react';
import store from '../store';
import { removeFromRoute } from '../actions/route';

export class RouteStop extends Component {
    remove(id) {
        store.dispatch(removeFromRoute(id));
    }

    render() {
        let sys = this.props.system;

        let classes = [
            'route-stop',
            (sys.RequestedStop) ? 'requested' : '',
        ];

        return (
            <li className={classes.join(' ') }>
                <div className="stop-details">
                    <h2>{sys.System.Name}</h2>
                    { sys.DistanceFromPrev > 0.01 ? (<p>{Math.floor(sys.DistanceFromPrev * 100) / 100} light years</p>) : (<p>Starting point</p>) }
                    { sys.System.ContainsScoopableStar ? (<p>Scoopable</p>) : (<p>Not scoopable</p>) }
                    <p></p>
                </div>

                { sys.RequestedStop ? (<button onClick={() => this.remove(sys.System.id)}><span className="btn-heading">Remove</span><span className="btn-detail">(-xx.xxly) </span></button>) : null }
            </li>
        );
    }
}