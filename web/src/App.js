import React, { Component } from 'react';
import './App.css';

import { SystemPicker } from './components/SystemPicker';
import { SystemHeading } from './components/SystemHeading';
import { RouteList } from './components/RouteList';
import { addToRoute } from './actions/route';

class App extends Component {
  add(id) {
    this.props.state.dispatch(addToRoute(id));
  }

  render() {
    return (
      <div className="App">
        <header>
          <h1>ED Paths</h1>
        </header>
        <main>
          <SystemPicker onActivate={id => this.add(id)} />
          <section className="route-overview">
            <SystemHeading system={this.props.state.getState().origin} />
            <p>{this.props.state.getState().distance}ly to</p>
            <SystemHeading system={this.props.state.getState().destination} />
          </section>
          <RouteList stops={this.props.state.getState().stops}></RouteList>
        </main>
      </div>
    );
  }
}

export default App;
