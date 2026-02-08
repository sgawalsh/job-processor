import { Routes, Route } from 'react-router-dom';
import JobForm from './jobForm';
import JobStatus from './jobStatus';
import NotFound from './notFound';

function App() {
  return (
    <Routes>
      <Route path="/" element={<JobForm />} />
      <Route path="/jobs/:id" element={<JobStatus />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}

export default App;