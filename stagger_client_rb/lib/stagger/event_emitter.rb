module Stagger
  module EventEmitter
    def on(event, callable = nil, &proc)
      _listeners[event] << (callable || proc)
    end

    def emit(event, *args)
      _listeners[event].each { |l| l.call(*args) }
    end

    def remove_listener(event, callable = nil, &proc)
      _listeners[event].delete(callable || proc)
    end

    def remove_all_listeners(event)
      _listeners.delete(event)
    end

    def listeners(event)
      _listeners[event]
    end

    private

    def _listeners
      @_listeners ||= Hash.new { |h,k| h[k] = [] }
    end
  end
end
