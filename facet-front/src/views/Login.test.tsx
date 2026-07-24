import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Login from './Login';

const mockLoginApi = vi.fn();
const mockNavigate = vi.fn();

vi.mock('@/api/auth', () => ({
  getCaptchaApi: vi.fn(() => Promise.resolve({ data: { captchaId: 'abc', base64Image: 'img' } })),
  loginApi: (...args: unknown[]) => mockLoginApi(...args),
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const mockStoreLogin = vi.fn();
vi.mock('@/stores/authStore', () => ({
  useAuthStore: vi.fn((selector) => {
    const store = { login: mockStoreLogin };
    return selector ? selector(store) : store;
  }),
}));

describe('Login', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders login form', () => {
    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );
    expect(screen.getByText('Facet')).toBeTruthy();
    expect(screen.getByText('Sign In')).toBeTruthy();
    expect(screen.getByLabelText('Email')).toBeTruthy();
    expect(screen.getByLabelText('Password')).toBeTruthy();
    expect(screen.getByLabelText('Captcha')).toBeTruthy();
  });

  it('shows error when fields are empty', async () => {
    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );
    fireEvent.click(screen.getByText('Sign In'));
    expect(await screen.findByText('All fields are required')).toBeTruthy();
  });

  it('calls loginApi and navigates on success', async () => {
    mockLoginApi.mockResolvedValue({ data: { token: 't1', user: { id: 'u1', email: 'a@b.com', displayName: 'A', role: 'admin' } } });

    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    await act(async () => {});
    await screen.findByAltText('Captcha');

    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'a@b.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'pwd' } });
    fireEvent.change(screen.getByLabelText('Captcha'), { target: { value: '1234' } });
    fireEvent.click(screen.getByText('Sign In'));

    await waitFor(() => {
      expect(mockLoginApi).toHaveBeenCalledWith({ email: 'a@b.com', password: 'pwd', captchaId: 'abc', captchaAnswer: '1234' });
    });
    expect(mockStoreLogin).toHaveBeenCalledWith('t1', { id: 'u1', email: 'a@b.com', displayName: 'A', role: 'admin' });
    expect(mockNavigate).toHaveBeenCalledWith('/overview', { replace: true });
  });

  it('shows error message on login failure', async () => {
    mockLoginApi.mockRejectedValue(new Error('Invalid credentials'));

    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    await act(async () => {});
    await screen.findByAltText('Captcha');

    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'a@b.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'bad' } });
    fireEvent.change(screen.getByLabelText('Captcha'), { target: { value: '1234' } });
    fireEvent.click(screen.getByText('Sign In'));

    expect(await screen.findByText('Invalid credentials')).toBeTruthy();
  });

  it('shows loading spinner on sign in button while loading', async () => {
    mockLoginApi.mockReturnValue(new Promise(() => {}));

    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    await act(async () => {});
    await screen.findByAltText('Captcha');

    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'a@b.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'pwd' } });
    fireEvent.change(screen.getByLabelText('Captcha'), { target: { value: '1234' } });
    fireEvent.click(screen.getByText('Sign In'));

    expect(document.querySelector('.MuiCircularProgress-root')).toBeTruthy();
  });
});
