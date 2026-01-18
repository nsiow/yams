import { Routes, Route } from 'react-router-dom';
import { Layout } from './components/layout';
import { HomePage } from './pages/home';
import { PreviewPage } from './pages/preview';

export function App(): JSX.Element {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/preview" element={<PreviewPage />} />
      </Route>
    </Routes>
  );
}
