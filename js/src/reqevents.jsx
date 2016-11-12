/* eslint-disable react/forbid-prop-types, no-console */
import 'babel-polyfill';
import React from 'react';
import ReactDOM from 'react-dom';


function createReqEventsURL() {
  const path = new URL(window.location).pathname;

  const m = path.match(/^\/g\/([^/]+)\/inspect/);
  const gomibakoKey = m[1];
  return `/g/${gomibakoKey}/reqevents`;
}
const reqEventsURL = createReqEventsURL();

function formatTimestamp(t) {
  return `${t.getFullYear()}-${t.getMonth() + 1}-${t.getDate()} ${t.getHours()}:${t.getMinutes()}`;
}

const Request = (props) => {
  const request = props.request;
  const headers = request.headers.map(h => (
    <tr key={h.key}><th>{h.key}</th><td>{h.value}</td></tr>
  ));

  return (
    <div>
      <div>{formatTimestamp(request.timestamp)}</div>
      <div>{request.method} {request.url}</div>
      <table>
        <tbody>{headers}</tbody>
      </table>
      <pre>{request.body}</pre>
    </div>
  );
};
Request.propTypes = {
  request: React.PropTypes.object,
};

class Requests extends React.Component {
  constructor() {
    super();
    this.state = {
      requests: [],
    };
  }
  componentDidMount() {
    const reqevents = new EventSource(reqEventsURL);
    reqevents.onmessage = (e) => {
      const r = JSON.parse(e.data);
      r.timestamp = new Date(r.timestamp * 1000);
      this.state.requests.unshift(r);
      this.setState({
        requests: this.state.requests,
      });
    };
  }
  render() {
    const requests = this.state.requests.map(r => (
      <li key={r.timestamp}><Request request={r} /></li>
    ));
    return (
      <ul>
        {requests}
      </ul>
    );
  }
}

Requests.propTypes = {
  requests: React.PropTypes.arrayOf(React.PropTypes.element),
};

ReactDOM.render(<Requests />, document.querySelector('#requests'));

