package com.tracker.client.activities.signin;

import java.util.logging.Logger;

import com.tracker.client.AppFactory;
import com.tracker.client.activities.BaseActivity;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.StyleUtils;

import elemental.dom.Element;

public class SignInActivity extends BaseActivity {
  private static final Logger log = Logger.getLogger(SignInActivity.class.getName());
  private FormSignIn formSignIn;
  private FormSignUp formSignUp;
  private Element formEl;

  public SignInActivity(AppFactory factory) {
    super(factory);
    formSignIn = new FormSignIn(factory);
    formSignUp = new FormSignUp(factory);
  }

  @Override
  protected void createDom() {
    decorateInternal(
        DomUtils.createDom(doc.createDivElement(), "container sign-in-up-container", 
          formEl = DomUtils.createDom(doc.createDivElement(), "row"))
    );
  }

  public Element getContentElement() {
    return formEl;
  }

  @Override
  public void start(StartCallback callback) {
    if (factory.getCurrentUser() != null) {
      factory.getPlaceController().goTo(factory.getNextPlace());
    } else {
      callback.start();
    }
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);

    addChild(formSignIn, true);
    addChild(formSignUp, true);

    StyleUtils.addClassName(formSignUp.getElement(), "border-left-md");
  }
}
