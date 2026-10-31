import { Component, ErrorInfo, ReactNode } from 'react';
import { AlertTriangle, RefreshCw, RotateCcw } from 'lucide-react';

interface Props {
  children: ReactNode;
  fallbackTitle?: string;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
    errorInfo: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error,
      errorInfo: null,
    };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught error rendered in dashboard:', error, errorInfo);
    this.setState({
      error,
      errorInfo,
    });
  }

  private handleRetry = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  private handleFullReload = () => {
    window.location.reload();
  };

  private handleClearAndReload = () => {
    try {
      localStorage.removeItem('vps_panel_token');
      sessionStorage.clear();
    } catch {
      // Ignore storage errors
    }
    window.location.reload();
  };

  public render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-slate-900 text-slate-100 flex items-center justify-center p-6 select-none">
          <div className="max-w-md w-full bg-slate-800/80 border border-slate-700/80 rounded-xl p-6 shadow-xl backdrop-blur-sm">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-lg bg-rose-500/10 border border-rose-500/20 flex items-center justify-center text-rose-400">
                <AlertTriangle className="w-5 h-5" />
              </div>
              <div>
                <h1 className="text-base font-semibold text-slate-100">
                  {this.props.fallbackTitle || 'Dashboard Interface Error'}
                </h1>
                <p className="text-xs text-slate-400">
                  A component encountered an unhandled rendering error.
                </p>
              </div>
            </div>

            {this.state.error && (
              <div className="mb-5 p-3 rounded-md bg-slate-950/60 border border-slate-800 text-xs font-mono text-rose-300 break-words max-h-36 overflow-y-auto">
                <div className="font-semibold text-rose-400 mb-1">
                  {this.state.error.name}: {this.state.error.message}
                </div>
                {this.state.error.stack && (
                  <div className="text-[11px] text-slate-500 whitespace-pre-wrap leading-relaxed">
                    {this.state.error.stack.split('\n').slice(1, 4).join('\n')}
                  </div>
                )}
              </div>
            )}

            <div className="flex items-center gap-2 pt-2 border-t border-slate-700/60">
              <button
                onClick={this.handleRetry}
                className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-xs font-medium rounded-lg bg-blue-600 hover:bg-blue-500 text-white transition-colors shadow-sm"
              >
                <RotateCcw className="w-3.5 h-3.5" />
                <span>Try Again</span>
              </button>

              <button
                onClick={this.handleFullReload}
                className="flex items-center justify-center gap-1.5 px-3 py-2 text-xs font-medium rounded-lg border border-slate-700 bg-slate-800 hover:bg-slate-700 text-slate-200 transition-colors"
                title="Reload page"
              >
                <RefreshCw className="w-3.5 h-3.5" />
                <span>Reload</span>
              </button>

              <button
                onClick={this.handleClearAndReload}
                className="px-3 py-2 text-xs font-medium rounded-lg border border-slate-700/60 bg-transparent hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition-colors"
                title="Reset session and reload"
              >
                Reset
              </button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
