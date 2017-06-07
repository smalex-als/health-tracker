package com.tracker.client.activities.signin;

import com.tracker.client.AppFactory;
import com.tracker.client.activities.widgets.FormButton;
import com.tracker.client.activities.widgets.FormInputText;
import com.tracker.client.jso.FormSubmitResp;
import com.tracker.client.jso.SignInReq;

import elemental.dom.Element;

public class FormSignIn extends BaseForm {
  private FormInputText emailEl = new FormInputText()
    .setName("username")
    .setTitle("Email or Username")
    .setPlaceholder("Enter email or username")
    .setSpellcheck(false)
    .setType("username");
  private FormInputText passwordEl = new FormInputText()
    .setName("password")
    .setTitle("Password")
    .setPlaceholder("Enter password")
    .setSpellcheck(false)
    .setType("password");
  private FormButton submitEl = new FormButton()
    .setTitle("Sign in")
    .setName("signin");

  public FormSignIn(AppFactory factory) {
    super(factory);
    setTitle("Welcome Back!");
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);

    addChild(emailEl, true);
    addChild(passwordEl, true);
    addChild(submitEl, true);
  }

  @Override
  public void enterDocument() {
    super.enterDocument();

    submitEl.onClick(() -> handleSubmit());
  }

  public void handleSubmit() {
    SignInReq in = SignInReq.create();
    in.setUsername(emailEl.getValue());
    in.setPassword(passwordEl.getValue());
    rpc.request("/v1/users/signin/", in, 
        (FormSubmitResp out) -> handleSubmitResp(out));
  }
}
