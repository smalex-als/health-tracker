package com.tracker.client.activities.widgets;

import com.tracker.client.controls.Component;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.StringUtils;
import com.tracker.client.util.StyleUtils;
import com.google.gwt.user.client.DOM;

import elemental.dom.Element;
import elemental.html.InputElement;
import elemental.html.LabelElement;
import java.util.logging.Logger;


public class FormCheckBox extends Component {
  private static final Logger log = Logger.getLogger(FormCheckBox.class.getName());
  private InputElement input;
  private LabelElement label;
  private String title = "";
  private String name = "";
  private String defaultValue;

  @Override
  protected void createDom() {
    input = DomUtils.createDom(doc.createInputElement(), "form-check-input");
    input.setType("checkbox");
    input.setName(name);
    input.setId(DOM.createUniqueId());
    input.setValue("true");
    label = DomUtils.createDom(doc.createLabelElement(), "form-check-label", input);
    label.appendChild(doc.createTextNode(" " + title));
    label.setHtmlFor(input.getId());
    decorateInternal(DomUtils.createDom(doc.createDivElement(), "form-group", label));
  }

  @Override
  protected void decorateInternal(Element element) {
    super.decorateInternal(element);
  }

  public String getValue() {
    return StringUtils.trimToEmpty(input.getValue());
  }

  public void setChecked(boolean checked) {
    input.setChecked(checked);
  }

  public boolean getChecked() {
    return input.isChecked();
  }

  public void setValue(String value) {
    input.setValue(StringUtils.trimToEmpty(value));
  }

  public FormCheckBox setTitle(String title) {
    this.title = title;
    return this;
  }

  public FormCheckBox setHasDanger(boolean value) {
    StyleUtils.toggleClass(getElement(), "has-danger", value);
    StyleUtils.hasClassName(getElement(), "has-danger");
    return this;
  }

  public boolean getHasDanger() {
    return StyleUtils.hasClassName(getElement(), "has-danger");
  }

  public FormCheckBox setDefaultValue(String defaultValue) {
    this.defaultValue = defaultValue;
    return this;
  }

  public FormCheckBox setName(String name) {
    this.name = name;
    return this;
  }

  public String getName() {
    return name;
  }
}
