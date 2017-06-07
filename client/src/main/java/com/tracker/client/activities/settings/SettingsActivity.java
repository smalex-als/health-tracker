package com.tracker.client.activities.settings;

import com.tracker.client.AppFactory;
import com.tracker.client.activities.BaseActivity;
import com.tracker.client.activities.widgets.FormButton;
import com.tracker.client.activities.widgets.FormInputText;
import com.tracker.client.activities.widgets.PopupMessage;
import com.tracker.client.rpc.ContentRpcService;
import com.tracker.client.rpc.ContentRpcService.AsyncCallback;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.FormUtils;
import com.tracker.client.util.NumberUtils;
import com.tracker.client.util.StyleUtils;
import com.google.gwt.core.client.JavaScriptObject;

import elemental.dom.Element;
import elemental.json.Json;
import elemental.json.JsonObject;

public class SettingsActivity extends BaseActivity {
  private Element formEl;
  private FormInputText numberCaloriesEl = new FormInputText()
    .setName("numberCalories")
    .setTitle("Expected number of calories per day")
    .setPlaceholder("Expected number of calories per day")
    .setSpellcheck(false)
    .setType("text");
  private FormButton saveEl = new FormButton()
    .setTitle("Save")
    .setName("save");
  protected ContentRpcService rpc;
  private final FormUtils formUtils;

  public SettingsActivity(AppFactory factory) {
    super(factory);
    this.rpc = factory.getRpcService();
    formUtils = new FormUtils(this);
  }

  @Override
  protected void createDom() {
    Element head = doc.createElement("h4");
    head.setClassName("form-settings-heading");
    head.setTextContent("Settings");

    decorateInternal(
        DomUtils.createDom(doc.createDivElement(), "container", 
          DomUtils.createDom(doc.createDivElement(), "row justify-content-center", 
            formEl = DomUtils.createDom(doc.createFormElement(), "form-settings-email", head)
            )
          )
        );
  }

  public Element getContentElement() {
    return formEl;
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);
    addChild(numberCaloriesEl, true);
    addChild(saveEl, true);
  }

  @Override
  public void enterDocument() {
    super.enterDocument();

    saveEl.onClick(() -> handleSubmit());
  }

  @Override
  public void start(StartCallback callback) {
    if (getElement() == null) {
      createDom();
    }
    rpc.request("GET", "/v1/users-settings/", null, new AsyncCallback<JsonObject>() {
      @Override
      public void onSuccess(JsonObject resp) {
        if (!formUtils.updateViewErrors(resp)) {
          updateView((int)resp.getNumber("numberCalories"));
        }
        callback.start();
      }
    });
  }

  private void updateView(int numberCalories) {
    formUtils.clearErrors();
    numberCaloriesEl.setValue(NumberUtils.toString(numberCalories));
  }
  
  public void handleSubmit() {
    StyleUtils.buttonEnable(saveEl.getElement(), false);
    JsonObject map = Json.createObject();
    map.put("numberCalories", NumberUtils.toInt(numberCaloriesEl.getValue()));
    JavaScriptObject jso = (JavaScriptObject) map.toNative();
    rpc.request("POST", "/v1/users-settings/", jso, new AsyncCallback<JsonObject>() {
      @Override
      public void onSuccess(JsonObject resp) {
        StyleUtils.buttonEnable(saveEl.getElement(), true);
        if (!formUtils.updateViewErrors(resp)) {
          addChild(new PopupMessage("Settings saved"), true);
        }
      }
    });
  }
}
