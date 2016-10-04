import React, { Component } from 'react';
import { RequestCompletion } from '../api/autocomplete';

export class SystemPicker extends Component {
    constructor() {
        super();

        this.state = { 
            fragment: '',
            options: [],
        };
    }

    add(id) {
        this.props.onActivate(id);
        this.setState({ fragment: '', options: [] });
    }

    autocomplete(event) {
        this.setState({
            fragment: event.target.value,
        }, () => {
            if (this.state.fragment.length > 2) {
                // Request autocompletes.
                RequestCompletion(this.state.fragment, results => {
                    this.setState({ options: results });
                });
            } else {
                this.setState({ options: [] });
            }
        })
    }

    render() {
        this.options = [{ id: 0, name: 'First' }, { id: 1, name: 'Second' }];

        let classes = [
            "system-options",
            (this.state.options.length === 0) ? "hide" : ""
        ];

        return (
            <div className="system-picker">
                <input value={this.state.fragment} onChange={event => this.autocomplete(event) } type="text" placeholder="Choose system..." />
                <ul className={classes.join(' ')}>{ this.state.options.map(sys => (<li onClick={() => this.add(sys.ID) } key={sys.ID}>{sys.Name}</li>)) }</ul>
            </div>
        );
    }
}