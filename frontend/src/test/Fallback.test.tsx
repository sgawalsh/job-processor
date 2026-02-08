import { render, screen } from '@testing-library/react';
import App from '../src/App';
import { expect, test } from 'vitest';
import { MemoryRouter } from 'react-router-dom';

test('unknown route shows fallback or redirects', () => {
  render(
    <MemoryRouter initialEntries={['/does-not-exist']}>
      <App />
    </MemoryRouter>
  );
  
  expect(screen.getByText(/Page Not Found./i)).toBeTruthy();
});