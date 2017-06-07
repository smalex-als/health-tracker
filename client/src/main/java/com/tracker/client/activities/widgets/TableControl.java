package com.tracker.client.activities.widgets;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.logging.Logger;

import com.tracker.client.controls.Component;
import com.tracker.client.util.DomUtils;
import com.tracker.shared.TableTemplate;
import com.google.gwt.core.client.Scheduler;
import com.google.gwt.core.client.Scheduler.ScheduledCommand;

import elemental.dom.Element;
import elemental.dom.NodeList;
import elemental.events.Event;
import elemental.events.EventListener;
import elemental.html.InputElement;
import elemental.html.TableCellElement;
import elemental.html.TableRowElement;

public class TableControl extends Component  {
  private static final Logger log = Logger.getLogger(TableControl.class.getName());

  public interface ClickHandler {
    void handleClick(int row, int cell, String dataId);
  }

  public interface SelectHandler {
    void handleSelect(List<String> ids);
  }
  private final Map<String, Object> map;
  private final TableTemplate template = new TableTemplate();
  private ClickHandler clickHandler;
  private SelectHandler selectHandler;

  public TableControl(Map<String, Object> map) {
    this.map = map;
  }

  @Override
  protected void createDom() {
    String body = template.toString(template.renderTable(map));
    decorateInternal((Element) DomUtils.htmlToDocumentFragment_(doc, body));
  }

  @Override
  public void enterDocument() {
    super.enterDocument();

    addHandlerRegistration(getElement().addEventListener(Event.CLICK, new EventListener() {
      @Override
      public void handleEvent(Event evt) {
        if (clickHandler == null) {
          return;
        }
        Element target = (Element) evt.getTarget();
        if (target.getTagName().equals("TD")) {
          TableCellElement cell = (TableCellElement) target;
          Element parent = target.getParentElement();
          if (parent.getNodeName().equals("TR")) {
            TableRowElement row = (TableRowElement) parent;
            evt.stopPropagation();
            evt.preventDefault();
            if (cell.getCellIndex() > 0) {
              String dataId = row.getAttribute("data-id");
              clickHandler.handleClick(row.getRowIndex(), cell.getCellIndex(), dataId);
            }
          }
        } else if (target.getTagName().equals("INPUT")) {
          if (selectHandler != null) {
            Scheduler.get().scheduleDeferred(new ScheduledCommand() {
              @Override
              public void execute() {
                selectHandler.handleSelect(getSelectedIds());
              }
            });
          }
        }
      }
    }, false));
  }

  public void onClick(ClickHandler handler) {
    clickHandler = handler;
  }

  public void onSelect(SelectHandler handler) {
    selectHandler = handler;
  }

  public List<String> getSelectedIds() {
    List<String> ids = new ArrayList<>();
    NodeList nodes = getElement().getElementsByTagName("INPUT");
    if (nodes.length() > 0) {
      for (int i = 0; i < nodes.length(); i++) {
        InputElement el = (InputElement) nodes.item(i);
        if (el.isChecked()) {
          Element parentEl = el.getParentElement().getParentElement();
          ids.add(parentEl.getAttribute("data-id"));
        }
      }
    }
    return ids;
  }
}
