package com.tracker.client.activities;

import com.tracker.client.AppFactory;
import com.tracker.client.place.BasePlace;
import com.tracker.client.rpc.ContentRpcService;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.FormUtils;

import elemental.dom.Element;
import elemental.json.JsonObject;

public class BaseEditActivity extends BaseActivity {
  protected final ContentRpcService rpc;
  protected Element formEl;
  protected BasePlace place;
  private final FormUtils formUtils;

  public BaseEditActivity(AppFactory factory) {
    super(factory);
    rpc = factory.getRpcService();
    formUtils = new FormUtils(this);
  }

  @Override
  protected void createDom() {
    Element head = doc.createElement("h4");
    head.setClassName("form-signin-heading");
    head.setTextContent("Edit");
    decorateInternal(
        DomUtils.createDom(doc.createDivElement(), "container sign-in-up-container", 
           DomUtils.createDom(doc.createDivElement(), "row"),
            DomUtils.createDom(doc.createDivElement(), "col-md-12", 
              formEl = DomUtils.createDom(doc.createFormElement(), "form-signin", head)
          ))
    );
  }

  public Element getContentElement() {
    return formEl;
  }

  public void updateForPlace(BasePlace basePlace) {
    this.place = basePlace;
  }

  protected void clearErrors() {
    formUtils.clearErrors();
  }

  protected boolean updateViewErrors(JsonObject resp) {
    formUtils.clearErrors();
    return formUtils.updateViewErrors(resp);
  }
}
