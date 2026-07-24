type NavigateFn = (path: string, opts?: { replace?: boolean }) => void;

let navigateFn: NavigateFn = (path) => {
  window.location.href = path;
};

export const setNavigate = (fn: NavigateFn) => {
  navigateFn = fn;
};

export const navigate = (path: string, opts?: { replace?: boolean }) => {
  navigateFn(path, opts);
};
