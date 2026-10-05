import {createRoot} from 'react-dom/client';
import App from './App.tsx';
import {createNetworkService} from './services/createNetworkService';
import '@fontsource/plus-jakarta-sans/400.css';
import '@fontsource/plus-jakarta-sans/500.css';
import '@fontsource/plus-jakarta-sans/600.css';
import '@fontsource/plus-jakarta-sans/700.css';
import '@fontsource/jetbrains-mono/400.css';
import '@fontsource/jetbrains-mono/500.css';
import '@fontsource/jetbrains-mono/600.css';
import './index.css';

createNetworkService().then(({service, startupError}) => {
  createRoot(document.getElementById('root')!).render(<App service={service} startupError={startupError} />);
});
