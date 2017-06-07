package com.tracker.client.activities.widgets;

import com.tracker.client.controls.Component;
import com.tracker.client.util.DomUtils;
import com.google.gwt.core.client.Scheduler;

public class PopupMessage extends Component {
  private final String message;

  public PopupMessage(String message) {
    this.message = message;
  }

  @Override
  protected void createDom() {
    decorateInternal(
        DomUtils.createDom(doc.createDivElement(), "alert-fixed alert alert-success offset-sm-3 col-sm-6"));
    getElement().setInnerText(message);
  }

  @Override
  public void enterDocument() {
    super.enterDocument();

    Scheduler.get().scheduleFixedDelay(() -> {
      getParent().removeChild(PopupMessage.this);
      return false;
    }, 2000);
  }

  @Override
  public void exitDocument() {
    super.exitDocument();
  }
}
