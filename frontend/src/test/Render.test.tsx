import { render } from '@testing-library/react';
import App from '../src/App';
import { expect, test } from 'vitest';
import { MemoryRouter } from 'react-router-dom';

test('app renders without crashing', () => {
  render(
    <MemoryRouter>
        <App />
    </MemoryRouter>
    );
  expect(document.body).toBeTruthy();
});

