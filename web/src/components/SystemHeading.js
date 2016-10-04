import React, { Component } from 'react';

export class SystemHeading extends Component {
    render() {
        if (this.props.system === null) return (<div></div>);

        return (
            <div>
                <h2>{this.props.system.Name}</h2>
            </div>
        );
    }
}