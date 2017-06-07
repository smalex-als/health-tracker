package com.tracker.client.activities.signin;

import com.tracker.client.AppFactory;
import com.tracker.client.controls.Component;
import com.tracker.client.jso.FormSubmitResp;
import com.tracker.client.rpc.ContentRpcService;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.FormUtils;

import elemental.dom.Element;

public class BaseForm extends Component {
  private Element formEl;
  private String title;
  protected ContentRpcService rpc;
  private AppFactory factory;
  private FormUtils formUtils;

  public BaseForm(AppFactory factory) {
    rpc = factory.getRpcService();
    this.factory = factory;
    formUtils = new FormUtils(this);
  }

  @Override
  protected void createDom() {
    Element head = doc.createElement("h4");
    head.setClassName("form-signin-heading");
    head.setTextContent(title);

    decorateInternal(
        DomUtils.createDom(doc.createDivElement(), "col-md-6", 
          formEl = DomUtils.createDom(doc.createFormElement(), "form-signin", head)
        )
    );
  }

  public Element getContentElement() {
    return formEl;
  }

  protected void handleSubmitResp(FormSubmitResp resp) {
    if (!formUtils.updateViewErrors(resp.getErrors())) {
      factory.setCurrentUser(resp.getUser());
      factory.getPlaceController().goTo(resp.getUser().getEmailConfirmed() 
          ? factory.getNextPlace() : factory.getConfirmEmailPlace());
    }
  }

  public BaseForm setTitle(String title) {
    this.title = title;
    return this;
  }
}
