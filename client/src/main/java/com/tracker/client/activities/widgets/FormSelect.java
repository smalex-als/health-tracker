package com.tracker.client.activities.widgets;

import java.util.HashMap;
import java.util.Map;
import java.util.logging.Logger;

import com.tracker.client.controls.Component;
import com.tracker.client.jso.ListBoxOption;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.StringUtils;
import com.tracker.client.util.StyleUtils;
import com.google.gwt.core.client.JsArray;
import com.google.gwt.user.client.DOM;

import elemental.dom.Element;
import elemental.html.LabelElement;
import elemental.html.OptionElement;
import elemental.html.SelectElement;

public class FormSelect extends Component {
  private static final Logger log = Logger.getLogger(FormSelect.class.getName());
  private Element rootEl;
  private SelectElement input;
  private LabelElement label;
  private Element feedbackEl;
  private String title = "";
  private String name = "";
  private boolean required = false;
  private Map<String, Integer> indexByValue = new HashMap<>();
  private JsArray<ListBoxOption> options;

  public FormSelect() {
  }

  @Override
  protected void createDom() {
    label = DomUtils.createDom(doc.createLabelElement(), "col-form-label");
    label.setTextContent(title);
    
    input = DomUtils.createDom(doc.createSelectElement(), "form-control");
    input.setId(DOM.createUniqueId());

    input.setName(name);
    label.setHtmlFor(input.getId());

    rootEl = DomUtils.createDom(doc.createDivElement(), "form-group",
        label,
        input
    );

    decorateInternal(rootEl);
    feedbackEl = DomUtils.createDom(doc.createDivElement(), "form-control-feedback");
    if (options != null) {
      for (int i = 0; i < options.length(); i++) {
        ListBoxOption option = options.get(i);
        addOption(option.getValue(), option.getName());
      }
    }
  }

  public void addOption(String value, String text) {
    indexByValue.put(value, input.getOptions().length());
    OptionElement opEl = doc.createOptionElement();
    opEl.setText(text);
    opEl.setValue(value);
    input.add(opEl, null);
  }

  @Override
  protected void decorateInternal(Element element) {
    super.decorateInternal(element);
  }

  public String getValue() {
    return StringUtils.trimToEmpty(input.getValue());
  }

  public void setValue(String value) {
    input.setValue(StringUtils.trimToEmpty(value));
  }

  public FormSelect setTitle(String title) {
    this.title = title;
    return this;
  }

  public FormSelect setOptions(JsArray<ListBoxOption> options) {
    this.options = options;
    return this;
  }

  public FormSelect setHasDanger(boolean value) {
    StyleUtils.toggleClass(rootEl, "has-danger", value);
    StyleUtils.hasClassName(rootEl, "has-danger");
    return this;
  }

  public boolean getHasDanger() {
    return StyleUtils.hasClassName(rootEl, "has-danger");
  }

  public FormSelect setRequired(boolean required) {
    this.required = required;
    return this;
  }

  public FormSelect setName(String name) {
    this.name = name;
    return this;
  }

  public String getName() {
    return name;
  }

  public void removeControlFeedback() {
    if (feedbackEl != null) {
      if (rootEl.contains(feedbackEl)) {
        rootEl.removeChild(feedbackEl);
      }
    }
  }

  public void setControlFeedback(String message) {
    feedbackEl.setTextContent(message);
    rootEl.appendChild(feedbackEl);
  }

  @Override
  public void enterDocument() {
    super.enterDocument();
  }

  public void setSelectedValue(String value) {
    if (indexByValue.containsKey(value)) {
      input.setSelectedIndex(indexByValue.get(value));
    } else {
      input.setSelectedIndex(0);
    }
  }

  public String getSelectedValue() {
    OptionElement el = (OptionElement) input.getOptions().item(input.getSelectedIndex());
    String val = el.getValue();
    return val;
  }
}
