import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from '../src/App';
import { expect, test } from 'vitest';
import { MemoryRouter } from 'react-router-dom';

test('user can create a job and see job status', async () => {
  render(
    <MemoryRouter>
        <App />
    </MemoryRouter>
    );

  const user = userEvent.setup();
  
  // Click submit
  const button = screen.getByRole('button', { name: /submit job/i });
  await user.click(button);

  // Assert job appears
  expect(
    await screen.queryByText('Description: My test job')
  ).toBeNull();
});