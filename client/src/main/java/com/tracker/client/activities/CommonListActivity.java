package com.tracker.client.activities;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import com.tracker.client.AppFactory;
import com.tracker.client.activities.widgets.PopupMessage;
import com.tracker.client.activities.widgets.TableControl;
import com.tracker.client.place.BasePlace;
import com.tracker.client.place.ViewId;
import com.tracker.client.rpc.ContentRpcService;
import com.tracker.client.rpc.ContentRpcService.AsyncCallback;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.StyleUtils;
import com.tracker.shared.Jsons;
import com.tracker.shared.TableTemplate;
import com.google.gwt.core.client.JavaScriptObject;
import com.google.gwt.user.client.Window;

import elemental.dom.Element;
import elemental.events.Event;
import elemental.events.EventListener;
import elemental.json.Json;
import elemental.json.JsonArray;
import elemental.json.JsonObject;
import java.util.logging.Logger;


public class CommonListActivity extends BaseActivity implements SearchComponent.Presenter {
  private static final Logger log = Logger.getLogger(CommonListActivity.class.getName());
  protected final ContentRpcService rpc;
  private boolean running;
  private String apiPrefix;
  private String apiForm;
  private String editViewId;
  private SearchComponent searchComponent;
  private TableControl table;
  private BasePlace place;
  private Element addEl;
  private Element deleteEl;

  public CommonListActivity(AppFactory factory) {
    super(factory);
    rpc = factory.getRpcService();
  }

  public CommonListActivity setApiPrefix(String apiPrefix) {
    this.apiPrefix = apiPrefix;
    return this;
  }

  public CommonListActivity setApiForm(String apiForm) {
    this.apiForm = apiForm;
    return this;
  }

  public CommonListActivity setEditViewId(String editViewId) {
    this.editViewId = editViewId;
    return this;
  }

  public CommonListActivity setSearchComponent(SearchComponent searchComponent) {
    this.searchComponent = searchComponent;
    searchComponent.setPresenter(this);
    return this;
  }

  @Override
  protected void createDom() {
    Map<String, Object> map = new HashMap<>();
    TableTemplate template = new TableTemplate();
    String body = template.toString(template.renderListBody(map));
    decorateInternal((Element) DomUtils.htmlToDocumentFragment_(doc, body));
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);
    addEl = getElementByClassName("list-btn-add");
    deleteEl = getElementByClassName("list-btn-rm");
    
    if (searchComponent != null) {
      Element parentEl = getElementByClassNameRequired("search-component");
      searchComponent.getComponent().render(parentEl);
    }
  }

  @Override
  public void enterDocument() {
    super.enterDocument();

    running = true;
    addHandlerRegistration(addEl.addEventListener(Event.CLICK, new EventListener() {
      @Override
      public void handleEvent(Event evt) {
        evt.stopPropagation();
        evt.preventDefault();
        handleOpen("new");
      }
    }, false));

    addHandlerRegistration(deleteEl.addEventListener(Event.CLICK, new EventListener() {
      @Override
      public void handleEvent(Event evt) {
        evt.stopPropagation();
        evt.preventDefault();
        handleDelete();
      }
    }, false));
  }

  @Override
  public void updateForPlace(BasePlace place) {
    this.place = place;
    if (running) {
      // reload for new place
      start(null);
    }
  }

  @Override
  public void exitDocument() {
    running = false;
    super.exitDocument();
  }

  public void start(StartCallback callback) {
    if (getElement() == null) {
      createDom();
    }
    JsonObject in = Json.createObject();
    Map<String, String> params = place.getParams();
    for (String key : params.keySet()) {
      in.put(key, params.get(key));
    }
    if (searchComponent != null) {
      searchComponent.updateView(in);
    }
    JavaScriptObject jso = (JavaScriptObject) in.toNative();
    rpc.request("GET", apiPrefix, jso, 
        (JsonObject out) -> handleResponse(out, callback));
  }

  private void handleDelete() {
    List<String> queue = table.getSelectedIds();
    if (queue.isEmpty()) {
      addChild(new PopupMessage("Nothing is selected"), true);
    } else if (Window.confirm("Are you sure?")) {
      processDeleteQueue(queue);
    }
  }

  private void processDeleteQueue(List<String> queue) {
    if (queue.isEmpty()) {
      start(null);
      addChild(new PopupMessage("Deleted"), true);
      return;
    }
    String id = queue.remove(queue.size() - 1);
    rpc.request("DELETE", apiPrefix + id, null , new AsyncCallback<JsonObject>() {
      @Override
      public void onSuccess(JsonObject resp) {
        JsonArray errors = resp.getArray("errors");
        if (errors != null && errors.length() > 0) {
          String message = errors.getObject(0).getString("message");
          addChild(new PopupMessage(message), true);
        } else {
          processDeleteQueue(queue);
        }
      }
    });
  }

  private void handleResponse(JsonObject resp, StartCallback callback) {
    if (table != null) {
      removeChild(table);
    }
    JsonArray errors = resp.getArray("errors");
    if (errors != null && errors.length() > 0) {
      String message = errors.getObject(0).getString("message");
      addChild(new PopupMessage(message), true);
    }
    table = new TableControl(Jsons.convertJsonObject(resp));
    addChild(table, true);
    table.onClick((int row, int cell, String id) -> handleOpen(id));
    table.onSelect((List<String> ids) -> handleSelect(ids));
    StyleUtils.addClassName(deleteEl, "disabled");
    
    if (callback != null) {
      callback.start();
    }
  }

  private void handleSelect(List<String> ids) {
    StyleUtils.toggleClass(deleteEl, "disabled", ids.size() == 0);
  }

  private void handleOpen(String id) {
    factory.getPlaceController().goTo(
        BasePlace.newBuilder()
        .parent(place)
        .viewId(new ViewId(editViewId)).id(id).build());
  }

  @Override
  public void clickSearch() {
    JsonObject values = searchComponent.updateModel();
    BasePlace.Builder builder = BasePlace.newBuilder()
        .viewId(new ViewId(place.getViewId().getActivityId()));
    for (String key : values.keys()) {
      builder.param(key, values.getString(key));
    }
    factory.getPlaceController().goTo(builder.build());
  }
}
