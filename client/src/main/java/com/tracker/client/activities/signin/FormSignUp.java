package com.tracker.client.activities.signin;

import com.tracker.client.AppFactory;
import com.tracker.client.activities.widgets.FormButton;
import com.tracker.client.activities.widgets.FormInputText;
import com.tracker.client.jso.FormSubmitResp;
import com.tracker.client.jso.SignUpReq;

import elemental.dom.Element;

public class FormSignUp extends BaseForm {
  private FormInputText emailEl = new FormInputText()
    .setName("email")
    .setTitle("Email")
    .setAutocomplete("off")
    .setPlaceholder("Enter email")
    .setSpellcheck(false)
    .setType("email");
  private FormInputText usernameEl = new FormInputText() 
    .setName("username")
    .setTitle("Username")
    .setPlaceholder("Enter username")
    .setAutocomplete("off")
    .setSpellcheck(false)
    .setType("username");
  private FormInputText passwordEl = new FormInputText()
    .setName("password")
    .setTitle("Password")
    .setPlaceholder("Enter password")
    .setType("password");
  private FormButton subminEl = new FormButton()
    .setTitle("Sign Up")
    .setName("signup");

  public FormSignUp(AppFactory factory) {
    super(factory);
    setTitle("New to tracker? Sign Up");
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);

    addChild(emailEl, true);
    addChild(usernameEl, true);
    addChild(passwordEl, true);
    addChild(subminEl, true);
  }

  @Override
  public void enterDocument() {
    super.enterDocument();

    subminEl.onClick(() -> handleSubmit());
  }

  public void handleSubmit() {
    SignUpReq in = SignUpReq.create();
    in.setEmail(emailEl.getValue());
    in.setUsername(usernameEl.getValue());
    in.setPassword(passwordEl.getValue());
    rpc.request("/v1/users/signup/", in, 
        (FormSubmitResp out) -> handleSubmitResp(out));
  }
}
