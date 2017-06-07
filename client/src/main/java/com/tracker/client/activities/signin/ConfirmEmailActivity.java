package com.tracker.client.activities.signin;

import com.tracker.client.AppFactory;
import com.tracker.client.activities.BaseActivity;
import com.tracker.client.activities.widgets.FormButton;
import com.tracker.client.activities.widgets.FormInputText;
import com.tracker.client.activities.widgets.FormLabel;
import com.tracker.client.activities.widgets.PopupMessage;
import com.tracker.client.jso.User;
import com.tracker.client.rpc.ContentRpcService;
import com.tracker.client.rpc.ContentRpcService.AsyncCallback;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.FormUtils;
import com.tracker.client.util.StyleUtils;

import elemental.dom.Element;
import elemental.json.JsonObject;

public class ConfirmEmailActivity extends BaseActivity {
  private Element formEl;
  private FormInputText codeEl = new FormInputText()
    .setName("code")
    .setTitle("Code")
    .setPlaceholder("Code")
    .setSpellcheck(false)
    .setType("text");
  private FormButton submitEl = new FormButton()
    .setTitle("Confirm")
    .setName("confirm");
  private FormButton sendConfirmationEl = new FormButton()
    .setTitle("Send confirmation again")
    .setName("sendConfirmation");
  protected ContentRpcService rpc;
  private final FormUtils formUtils;

  public ConfirmEmailActivity(AppFactory factory) {
    super(factory);
    this.rpc = factory.getRpcService();
    formUtils = new FormUtils(this);
  }

  @Override
  protected void createDom() {
    Element head = doc.createElement("h4");
    head.setClassName("form-confirm-email-heading");
    head.setTextContent("Confirm email");

    decorateInternal(
        DomUtils.createDom(doc.createDivElement(), "container", 
          DomUtils.createDom(doc.createDivElement(), "row justify-content-center", 
            formEl = DomUtils.createDom(doc.createFormElement(), "form-confirm-email", head)
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
    addChild(codeEl, true);
    addChild(submitEl, true);
    FormLabel formLabel = new FormLabel();
    formLabel.setText("<br/>Or you can send email confirmation again");
    addChild(formLabel, true);

    addChild(sendConfirmationEl, true);
  }

  @Override
  public void enterDocument() {
    super.enterDocument();

    submitEl.onClick(() -> handleSubmit());
    sendConfirmationEl.onClick(() -> handleSendConfirmation());
    codeEl.setValue("");
  }

  public void handleSubmit() {
    String value = codeEl.getValue();
    StyleUtils.buttonEnable(submitEl.getElement(), false);
    StyleUtils.buttonEnable(sendConfirmationEl.getElement(), false);
    rpc.request("GET", "/v1/users-confirm/?code=" + value, null, new AsyncCallback<JsonObject>() {
      @Override
      public void onSuccess(JsonObject resp) {
        StyleUtils.buttonEnable(submitEl.getElement(), true);
        StyleUtils.buttonEnable(sendConfirmationEl.getElement(), true);
        if (!formUtils.updateViewErrors(resp)) {
          addChild(new PopupMessage("Email confirmed"), true);
          User user = factory.getCurrentUser();
          user.setEmailConfirmed(true);
          factory.setCurrentUser(user);
          factory.getPlaceController().goTo(factory.getNextPlace());
        }
      }
    });
  }

  public void handleSendConfirmation() {
    StyleUtils.buttonEnable(submitEl.getElement(), false);
    StyleUtils.buttonEnable(sendConfirmationEl.getElement(), false);
    rpc.request("GET", "/v1/users-send-confirmation/", null, new AsyncCallback<JsonObject>() {
      @Override
      public void onSuccess(JsonObject resp) {
        StyleUtils.buttonEnable(submitEl.getElement(), true);
        StyleUtils.buttonEnable(sendConfirmationEl.getElement(), true);
        if (!formUtils.updateViewErrors(resp)) {
          addChild(new PopupMessage("Confirmation sent"), true);
        }
      }
    });
  }
}
