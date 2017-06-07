package com.tracker.client.rpc;

import java.util.Map;
import java.util.logging.Logger;

import com.tracker.client.AppFactory;
import com.tracker.shared.Jsons;
import com.google.gwt.core.client.JavaScriptObject;
import com.google.gwt.core.client.JsonUtils;
import com.google.gwt.core.client.Scheduler;
import com.google.gwt.core.client.Scheduler.RepeatingCommand;
import com.google.gwt.http.client.URL;
import com.google.gwt.user.client.Window;

import elemental.client.Browser;
import elemental.events.Event;
import elemental.events.EventListener;
import elemental.json.JsonObject;
import elemental.xml.XMLHttpRequest;

public class ContentRpcServiceImpl implements ContentRpcService {

  private static final Logger log = Logger.getLogger(ContentRpcServiceImpl.class.getName());

  /**
   * Encapsulates a linked list node that is used by {@link TaskQueue} to keep
   * an ordered list of pending {@link Task}s.
   */
  private static class Node {
    private final Task task;

    private Node next;

    Node(Task task) {
      this.task = task;
    }

    void execute(TaskQueue queue) {
      task.execute(queue);
    }
  }

  /**
   * Encapsulates a task for writing data to the server. The tasks are managed
   * by the {@link TaskQueue} and are auto-retried on failure.
   */
  private abstract static class Task {
    private TaskQueue queue;

    abstract void execute();

    void execute(TaskQueue queue) {
      this.queue = queue;
      execute();
    }

    TaskQueue getQueue() {
      return queue;
    }
  }

  /**
   * Provides a mechanism to perform write tasks sequentially and retry tasks
   * that fail.
   */
  private class TaskQueue extends RetryTimer {
    private Node head, tail;

    public void post(Task task) {
      final Node node = new Node(task);
      if (isIdle()) {
        head = tail = node;
        executeHead();
      } else {
        enqueueTail(node);
      }
    }

    private void enqueueTail(Node node) {
      assert head != null && tail != null;
      assert node != null;
      tail = tail.next = node;
    }

    private void executeHead() {
      head.execute(this);
    }

    private void executeNext() {
      head = head.next;
      if (head != null) {
        executeHead();
      } else {
        tail = null;
      }
    }

    private boolean isIdle() {
      return head == null;
    }

    private void taskFailed(Task task, boolean fatal) {
      assert task == head.task;

      // Report a failure to the Model.
      onServerFailed(fatal);

      // Schedule a retry.
      retryLater();
    }

    private void taskSucceeded(Task task) {
      assert task == head.task;
      // Report a success to the Model.
      onServerSucceeded();

      // Reset the retry counter.
      resetRetryCount();

      // Move on to the next task.
      executeNext();
    }

    @Override
    protected void retry() {
      // Retry running the head task.
      executeHead();
    }
  }

  private class RequestTask<T extends JavaScriptObject> extends Task {
    private final String method;
    private final String url;
    private final String data;
    private final AsyncCallback<T> asyncCallback;

    public RequestTask(String method, String url, String data, AsyncCallback<T> asyncCallback) {
      this.method = method;
      this.url = url;
      this.data = data;
      this.asyncCallback = asyncCallback;
    }

    @Override
    void execute() {
      final XMLHttpRequest xhr = Browser.getWindow().newXMLHttpRequest();
      xhr.setOnerror(new EventListener() {
        @Override
        public void handleEvent(Event evt) {
          onError("evt " + xhr.getStatus());
        }
      });
      xhr.setOnload(new EventListener() {
        @Override
        public void handleEvent(Event evt) {
          onLoad(xhr, evt);
        }
      });
      statusObserver.onTaskStarted("loading...");
      xhr.open(method, url);
      xhr.setRequestHeader("Content-Type", "application/json;charset=utf-8");
      if (data != null) {
        xhr.send(data);
      } else {
        xhr.send();
      }
    }

    private void onLoad(XMLHttpRequest xhr, Event evt) {
      statusObserver.onTaskFinished();

      if (xhr.getStatus() == 401) {
        getQueue().taskSucceeded(this);
        // appFactory.setCurrentUser(null);
        // appFactory.getPlaceController().goTo(appFactory.getDefaultPlace());
        Scheduler.get().scheduleFixedDelay(new RepeatingCommand() {
          @Override
          public boolean execute() {
            Window.Location.assign("/");
            // Window.Location.assign("#signin");
            return false;
          }
        }, 1000);
        return;
      }
      if (xhr.getStatus() != 200 && xhr.getStatus() != 400 && xhr.getStatus() != 403 && xhr.getStatus() != 404) {
        getQueue().taskFailed(this, false);
        return;
      }
      getQueue().taskSucceeded(this);
      T t = (T) JsonUtils.unsafeEval(xhr.getResponseText());
      asyncCallback.onSuccess(t);
    }

    private void onError(String msg) {
      getQueue().taskFailed(this, false);
      Window.alert("ERROR: " + msg);
    }
  }

  /**
   * A task queue to manage all writes to the server.
   */
  private final TaskQueue taskQueue = new TaskQueue();

  /**
   * Indicates whether the RPC end point is currently responding.
   */
  private boolean offline;

  /**
   * The observer that is receiving status events.
   */
  private final StatusObserver statusObserver;

  private final AppFactory appFactory;

  public ContentRpcServiceImpl(AppFactory appFactory, StatusObserver statusObserver) {
    this.appFactory = appFactory;
    this.statusObserver = statusObserver;
  }

  @Override
  public void request(String url, JavaScriptObject in, AsyncCallback asyncCallback) {
    request("POST", url, in, asyncCallback);
  }

  @Override
  public void request(String method, String url, JavaScriptObject in, AsyncCallback asyncCallback) {
    if (!offline) {
      // String data = in != null ? (new JSONObject(in).toString()) : null;
      String data = in != null ? JsonUtils.stringify(in) : null;
      if (in != null && method.equals("GET")) {
        JsonObject obj = JsonUtils.safeEval(data);
        Map<String, Object> params = Jsons.convertJsonObject(obj);
        if (!params.isEmpty()) {
          StringBuilder sb = new StringBuilder();
          sb.append(url);
          sb.append("?");
          for (String key : params.keySet()) {
            sb.append(key);
            sb.append("=");
            sb.append(URL.encodeQueryString(String.valueOf(params.get(key))));
            sb.append("&");
          }
          url = sb.toString();
        }
      }
      taskQueue.post(new RequestTask(method, url, data, asyncCallback));
    }
  }

  /**
   * Invoked by tasks and loaders when RPC invocations begin to fail.
   */
  void onServerFailed(boolean fatal) {
    if (fatal) {
      forceApplicationReload();
      return;
    }

    if (!offline) {
      statusObserver.onServerWentAway();
      offline = true;
    }
  }

  /**
   * Invoked by tasks and loaders when RPC invocations succeed.
   */
  void onServerSucceeded() {
    if (offline) {
      statusObserver.onServerCameBack();
      offline = false;
    }
  }

  static native void forceApplicationReload() /*-{
    $wnd.location.reload();
  }-*/;
}
