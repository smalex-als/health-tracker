package com.tracker.client.activities;

import java.util.logging.Logger;

import com.tracker.client.AppFactory;
import com.tracker.client.activities.widgets.FormButton;
import com.tracker.client.activities.widgets.PopupMessage;
import com.tracker.client.jso.EditFormResp;
import com.tracker.client.place.BasePlace;
import com.tracker.client.place.ViewId;
import com.tracker.client.rpc.ContentRpcService.AsyncCallback;
import com.tracker.client.util.NumberUtils;
import com.tracker.client.util.StyleUtils;
import com.google.gwt.core.client.JavaScriptObject;

import elemental.dom.Element;
import elemental.json.Json;
import elemental.json.JsonObject;

public class CommonEditActivity extends BaseEditActivity {
  private static final Logger log = Logger.getLogger(CommonEditActivity.class.getName());
  private FormButton saveEl = new FormButton()
    .setTitle("Save")
    .setName("save");
  private EditViewComponent editViewComponent;
  private JsonObject model;
  private String apiPrefix;
  private String apiForm;

  public CommonEditActivity(AppFactory factory) {
    super(factory);
  }

  public CommonEditActivity setApiPrefix(String apiPrefix) {
    this.apiPrefix = apiPrefix;
    return this;
  }

  public CommonEditActivity setApiForm(String apiForm) {
    this.apiForm = apiForm;
    return this;
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);
  }

  @Override
  public void enterDocument() {
    super.enterDocument();

    saveEl.onClick(() -> handleSave());
  }

  @Override
  public void start(StartCallback callback) {
    if (getElement() == null) {
      createDom();

      rpc.request("GET", apiForm, null, new AsyncCallback<EditFormResp>() {
        @Override
        public void onSuccess(EditFormResp t) {
          editViewComponent = new EditViewComponent(t);
          addChild(editViewComponent, true);
          addChild(saveEl, true);
          innerStart(callback);
        }
      });
    } else {
      innerStart(callback);
    }
  }

  public void innerStart(StartCallback callback) {
    if (place.getId().equals("new")) {
      updateView(null);
      callback.start();
    } else {
      rpc.request("GET", apiPrefix + place.getId(), null, new AsyncCallback<JsonObject>() {
        @Override
        public void onSuccess(JsonObject resp) {
          if (!updateViewErrors(resp)) {
            updateView(resp.getObject("data"));
          }
          callback.start();
        }
      });
    }
  }

  private void updateView(JsonObject model) {
    clearErrors();
    if (model == null) {
      model = Json.createObject();
    }
    this.model = model;
    editViewComponent.updateView(this.model);
  }

  private JsonObject updateModel() {
    editViewComponent.updateModel();
    return model;
  }

  private void handleSave() {
    JsonObject model = updateModel();
    if (model == null) {
      return;
    }
    JavaScriptObject in = (JavaScriptObject) model.toNative();
    StyleUtils.buttonEnable(saveEl.getElement(), false);
    rpc.request("POST", apiPrefix, in, new AsyncCallback<JsonObject>() {
      @Override
      public void onSuccess(JsonObject resp) {
        StyleUtils.buttonEnable(saveEl.getElement(), true);
        if (!updateViewErrors(resp)) {
          JsonObject data = resp.getObject("data");
          long newId = getObjectLong(data, "id");
          boolean reload = newId != getObjectLong(model, "id");
          addChild(new PopupMessage("Saved"), true);
          updateView(data);
          if (reload) {
            reloadForm(String.valueOf(newId));
          }
          // User curUser = factory.getCurrentUser();
          // if (resp.getUser().getId().equals(curUser.getId())) {
          //   factory.setCurrentUser(resp.getUser());
          // }
        }
      }
    });
  }

  public long getObjectLong(JsonObject obj, String field) {
    if (obj != null) {
      JsonObject value = obj.get(field);
      if (value != null) {
        return NumberUtils.toLong(value.asString());
      }
    }
    return 0L;
  }

  private void reloadForm(String id) {
    BasePlace.Builder builder = BasePlace.newBuilder();
    if (place.getParent() != null) {
      builder.parent(place.getParent());
    }
    String activityId = place.getViewId().getActivityId();
    builder.viewId(new ViewId(activityId)).id(id);
    factory.getPlaceController().goTo(builder.build());
  }
}
